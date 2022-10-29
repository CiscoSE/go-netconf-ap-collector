package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/openshift-telco/go-netconf-client/netconf"
	"github.com/openshift-telco/go-netconf-client/netconf/message"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/ssh"
)

func main() {
	// Load config file
	// read our opened jsonFile as a byte array.
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	configB, err := ioutil.ReadAll(configFile)

	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}

	var configuration Configuration
	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	err = json.Unmarshal(configB, &configuration)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	// Start MQTT client
	var broker = configuration.MQTTConfig.Broker
	var port = configuration.MQTTConfig.Port
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID(configuration.MQTTConfig.ClientId)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Start MongoDB connection

	c := make(chan string)
	for i := 0; i < len(configuration.Collection.WirelessControllers); i++ {
		wc := configuration.Collection.WirelessControllers[i]
		go func(wc *WirelessController, mqttc *mqtt.Client, c chan string) {
			notificationSession := createSession(wc.Name, int(wc.Port))

			callback := func(event netconf.Event) {
				reply := event.Notification()
				fmt.Printf("** Handling notification from %v **\n", wc.Name)

				var notification ApNotification
				xml.Unmarshal([]byte(reply.RawReply), &notification)
				mclient, mctx, _, _ := ConnectMongo(configuration.MongoConfig.Url)
				for i := 0; i < len(notification.PushUpdate.Content.ApGlobalOperData.ApJoinStats); i++ {
					apStats := notification.PushUpdate.Content.ApGlobalOperData.ApJoinStats[i]
					_, err := FindByMac(mclient, mctx, "wireless", "onboarded_aps", apStats.WtpMac)
					if err != nil {
						if fmt.Sprintf("%v", err) == "mongo: no documents in result" {

							fmt.Printf("\nNew AP found in  %v", apStats.WtpMac)
							fmt.Printf("\nIs Joined: %v", apStats.ApJoinInfo.IsJoined)
							mqttcObj := *mqttc
							mqttcObj.Publish(configuration.MQTTConfig.Topic, 0, false, "{\"mac\":\""+apStats.WtpMac+"\"}")
							document := bson.D{
								{Key: "mac", Value: apStats.WtpMac},
							}
							_, err := insertOne(mclient, mctx, "wireless",
								"onboarded_aps", document)

							if err != nil {
								fmt.Printf("Error inserting AP MAC: %v\n", err)
							}
						} else {
							fmt.Printf("Error searching AP MAC: %v\n", err)
						}
					}
				}
			}

			// TODO: This should be modeled with a struct
			createSubscription := `
			<establish-subscription
			xmlns="urn:ietf:params:xml:ns:yang:ietf-event-notifications"
			xmlns:yp="urn:ietf:params:xml:ns:yang:ietf-yang-push">
				<stream>yp:yang-push</stream>
				<yp:xpath-filter>/wireless-ap-global-oper:ap-global-oper-data/ap-join-stats/ap-join-info/is-joined</yp:xpath-filter>
				<yp:period>1000</yp:period>
			</establish-subscription>
			`
			rpc := message.NewRPC(createSubscription)
			reply, err := notificationSession.SyncRPC(rpc, 1)
			if err != nil {
				fmt.Printf("Error: %v", err)
			}
			fmt.Printf("** Subscription to %v created - ID %v **\n", wc.Name, reply.SubscriptionID)
			notificationSession.Listener.Register(reply.SubscriptionID, callback)
			notificationSession.Listener.WaitForMessages()
		}(&wc, &client, c)

	}
	for {
		<-c
	}

}

func createSession(ip string, port int) *netconf.Session {
	// TODO: get creds from environment
	sshConfig := &ssh.ClientConfig{
		User:            os.Getenv("WLC_USER"),
		Auth:            []ssh.AuthMethod{ssh.Password(os.Getenv("WLC_PASSWORD"))},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	s, err := netconf.DialSSH(fmt.Sprintf("%v:%d", ip, port), sshConfig)
	if err != nil {
		log.Fatal(err)
	}
	capabilities := netconf.DefaultCapabilities
	err = s.SendHello(&message.Hello{Capabilities: capabilities})
	if err != nil {
		log.Fatal(err)
	}

	return s
}
