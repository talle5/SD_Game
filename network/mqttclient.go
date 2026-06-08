package network

import (
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	brokerHost string = "broker.emqx.io"
	brokerPort int    = 1883
	clientID string = fmt.Sprintf("go-%d", time.Now().UnixNano())

	mqttMu     sync.Mutex
	mqttClient mqtt.Client
)

// MQTTClient retorna uma instância do cliente MQTT. Se já existir uma conexão ativa,
// ela será reutilizada (padrão Singleton). Utiliza um Mutex para garantir segurança em concorrência.
func MQTTClient() (mqtt.Client, error) {
	mqttMu.Lock()
	defer mqttMu.Unlock()

	if mqttClient != nil && mqttClient.IsConnected() {
		return mqttClient, nil
	}

	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s:%d", brokerHost, brokerPort)).
		SetClientID(clientID)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, token.Error()
	}

	mqttClient = client
	return mqttClient, nil
}

// PublishMQTT publica uma mensagem em um tópico no broker MQTT,
// permitindo definir explicitamente se a mensagem deve ser retida (retained) ou não.
func PublishMQTT(topic string, payload []byte, retained bool) error {
	client, err := MQTTClient()
	if err != nil {
		return err
	}

	token := client.Publish(topic, 1, retained, payload)
	token.Wait()
	return token.Error()
}

// SubscribeMQTT inscreve o cliente em um tópico MQTT e define uma função de callback (handler)
// que será executada sempre que uma nova mensagem for recebida no tópico especificado.
func SubscribeMQTT(topic string, handler func([]byte)) error {
	client, err := MQTTClient()
	if err != nil {
		return err
	}

	token := client.Subscribe(topic, 1, func(_ mqtt.Client, msg mqtt.Message) {
		handler(msg.Payload())
	})
	token.Wait()
	return token.Error()
}

func GetClientID() string {
	return clientID
}