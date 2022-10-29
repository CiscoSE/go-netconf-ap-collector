# NETCONF collector for new access points

This aplication is an example of how you can use Go to establish a subscription to detect new access points joined to a controller, and integrate that to MQTT. Since the wireless controller returns all access points, a Mongo database is used to keep track of what has been detected. Only new access points trigger a message to the MQTT broker.

# Installation

1. Clone the repo
2. Create a file named config.json - example content:

```json
{
    "collection": {
        "wireless-controllers": [
            {
                "name": "192.168.6.1",
                "port": 830
            }
        ]
    },
    "mqtt": {
        "broker": "mosquitto.example.com",
        "port": 1883,
        "client-id": "go_mqtt_client",
        "topic":"wireless/ap"
    },
    "mongo": {
        "url":"mongodb://mongo.example.com:27017"
    }
}
```

3. Add credentials to enviromental variables:

```bash
export WLC_USER=admin
export WLC_PASSWORD=supersecret
```

4. Build and execute

```bash
netconf-collector % go build
netconf-collector % ./netconf-collector
Connected to MQTT
<?xml version="1.0" encoding="UTF-8"?>
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0"><capabilities><capability>urn:ietf:params:netconf:base:1.0</capability><capability>urn:ietf:params:netconf:base:1.1</capability></capabilities></hello>]]>]]>


Sending RPC

#517
<?xml version="1.0" encoding="UTF-8"?>
<rpc xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="333df2a7-c18a-4460-86af-99beba990407">
			<establish-subscription
			xmlns="urn:ietf:params:xml:ns:yang:ietf-event-notifications"
			xmlns:yp="urn:ietf:params:xml:ns:yang:ietf-yang-push">
				<stream>yp:yang-push</stream>
				<yp:xpath-filter>/wireless-ap-global-oper:ap-global-oper-data/ap-join-stats/ap-join-info/is-joined</yp:xpath-filter>
				<yp:period>1000</yp:period>
			</establish-subscription>
			</rpc>
##

Successfully executed RPC
** Subscription to 192.168.6.1 created - ID 2147483667 **
** Handling notification from 192.168.6.1 **
^C

```

