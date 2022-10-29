package main

type Configuration struct {
	Collection  Collection `json:"collection"`
	MQTTConfig  MQTTConfig `json:"mqtt"`
	MongoConfig Mongo      `json:"mongo"`
}

type Collection struct {
	WirelessControllers []WirelessController `json:"wireless-controllers"`
}
type WirelessController struct {
	Name string `json:"name"`
	Port int32  `json:"port"`
}

type MQTTConfig struct {
	Broker   string `json:"broker"`
	Port     int32  `json:"port"`
	ClientId string `json:"client-id"`
	Topic    string `json:"topic"`
}

type Mongo struct {
	Url string `json:"url"`
}
