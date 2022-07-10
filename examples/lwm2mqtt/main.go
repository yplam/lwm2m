package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/yplam/lwm2m"
)

var (
	brokerAddr = "tcp://192.168.2.235:31883"
)

type Manager struct {
	cli mqtt.Client
}

type Device struct {
	Name            string   `json:"name"`
	Model           string   `json:"model"`
	Manufacturer    string   `json:"manufacturer"`
	SoftwareVersion string   `json:"sw_version"`
	Identifiers     []string `json:"identifiers"`
}

type ConfigMessage struct {
	Name              string `json:"name"`
	StateTopic        string `json:"state_topic"`
	StateClass        string `json:"state_class"`
	DeviceClass       string `json:"device_class"`
	UnitOfMeasurement string `json:"unit_of_measurement"`
	Device            Device `json:"device"`
	ValueTemplate     string `json:"value_template"`
	UniqueID          string `json:"unique_id"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logrus.Info("starting lwm2m server")
	m := &Manager{}
	go m.Serve(ctx)
	s := lwm2m.NewServer(
		lwm2m.WithOnNewDeviceConn(m.OnLWM2MDeviceConn),
		lwm2m.EnableUDPListener("udp6", ":5683"),
		lwm2m.EnableDTLSListener("udp6", ":5684", lwm2m.NewDummy()))
	go s.Serve(ctx)
	go udpMulticast(ctx)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	cancel()
	logrus.Info("Shutting down.")
}

func (m *Manager) OnLWM2MDeviceConn(d *lwm2m.Device) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	logrus.Infof("new device: %s", d.EndPoint)
	var manufacturer, model, name, version string
	if d.HasObjectInstance(3, 0) {
		p, _ := lwm2m.NewPathFromString("3/0")
		nodes, err := d.Read(ctx, p)
		if err != nil {
			logrus.Warnf("read device info error %v", err)
			return
		}
		if r, err := lwm2m.NodeGetResourceByPath(nodes,
			lwm2m.NewResourcePath(3, 0, 0)); err == nil {
			manufacturer = r.StringValue()
		}
		if r, err := lwm2m.NodeGetResourceByPath(nodes,
			lwm2m.NewResourcePath(3, 0, 1)); err == nil {
			model = r.StringValue()
		}
		if r, err := lwm2m.NodeGetResourceByPath(nodes,
			lwm2m.NewResourcePath(3, 0, 2)); err == nil {
			name = r.StringValue()
		}
		if r, err := lwm2m.NodeGetResourceByPath(nodes,
			lwm2m.NewResourcePath(3, 0, 3)); err == nil {
			version = r.StringValue()
		}

		pr, _ := lwm2m.NewPathFromString("3/0/2")
		_, err = d.Read(ctx, pr)
		if err != nil {
			logrus.Warnf("read device euid error %v", err)
			return
		} else {
			logrus.Infof("read euid ok")
		}
	}
	di := Device{
		Name:            name,
		Model:           model,
		Manufacturer:    manufacturer,
		SoftwareVersion: version,
		Identifiers: []string{
			d.EndPoint,
		},
	}
	logrus.Info("read info ok")
	if d.HasObjectWithInstance(3300) {
		topic := fmt.Sprintf("homeassistant/sensor/%v/lwm2mqtt_3300_0_5700/config", d.EndPoint)
		stateTopic := fmt.Sprintf("lwm2mqtt/%v", d.EndPoint)
		msg := ConfigMessage{
			Name:              "CO2 Sensor",
			StateTopic:        stateTopic,
			DeviceClass:       "carbon_dioxide",
			UnitOfMeasurement: "ppm",
			StateClass:        "measurement",
			Device:            di,
			ValueTemplate:     "{{ value_json['3300_0_5700'] }}",
			UniqueID:          fmt.Sprintf("%v_lwm2mqtt_3300_0_5700", d.EndPoint),
		}
		data, _ := json.Marshal(msg)
		d.AddObservation(lwm2m.NewObjectPath(3300), m.OnLWM2MMessage)
		m.cli.Publish(topic, 1, true, data)
		logrus.Infof("public to %v, %v", topic, string(data))

	}
	if d.HasObjectWithInstance(3303) {
		topic := fmt.Sprintf("homeassistant/sensor/%v/lwm2mqtt_3303_0_5700/config", d.EndPoint)
		stateTopic := fmt.Sprintf("lwm2mqtt/%v", d.EndPoint)
		msg := ConfigMessage{
			Name:              "Temperature",
			StateTopic:        stateTopic,
			DeviceClass:       "temperature",
			UnitOfMeasurement: "°C",
			StateClass:        "measurement",
			Device:            di,
			ValueTemplate:     "{{ value_json['3303_0_5700'] }}",
			UniqueID:          fmt.Sprintf("%v_lwm2mqtt_3303_0_5700", d.EndPoint),
		}
		data, _ := json.Marshal(msg)
		d.AddObservation(lwm2m.NewObjectPath(3303), m.OnLWM2MMessage)
		m.cli.Publish(topic, 1, true, data)
		logrus.Infof("public to %v, %v", topic, string(data))

	}
}

func (m *Manager) OnLWM2MMessage(d *lwm2m.Device, p lwm2m.Path, notify []lwm2m.Node) {
	logrus.Infof("new notify from %s:%s", d.EndPoint, p.String())

	val, err := lwm2m.NodeGetAllResources(notify, p)
	if err != nil {
		logrus.Warnf("read val err (%v), (%v)", p.String(), err)
	} else {
		content := make(map[string]interface{})
		stateTopic := fmt.Sprintf("lwm2mqtt/%v", d.EndPoint)
		logrus.Infof("get val from (%v)", p)
		for k, v := range val {
			mk := strings.ReplaceAll(strings.TrimLeft(k.String(), "/"), "/", "_")
			vv := v.Value()
			if reflect.TypeOf(vv).Name() == "float64" {
				vv = math.Round(vv.(float64)*100) / 100
			}
			content[mk] = vv
		}
		data, _ := json.Marshal(content)
		m.cli.Publish(stateTopic, 1, true, data)
		logrus.Infof("public to %v, %v", stateTopic, string(data))
	}
}

func (m *Manager) OnMQTTMessage(client mqtt.Client, message mqtt.Message) {
	logrus.Infof("new mqtt message %v", message.Topic())
}

func (m *Manager) Serve(ctx context.Context) {
	mqtt.ERROR = logrus.New()
	opts := mqtt.NewClientOptions().
		AddBroker(brokerAddr).
		SetClientID("lwm2mqtt").
		SetUsername("lwm2mqtt").
		SetPassword("3u62Iw3u8h4x").
		SetKeepAlive(30 * time.Second).
		SetPingTimeout(5 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := c.Subscribe("homeassistant/+/lwm2mqtt/+/set", 1, m.OnMQTTMessage); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return
	}

	m.cli = c

	<-ctx.Done()

	if token := c.Unsubscribe("homeassistant/+/lwm2mqtt/+/set"); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return
	}
	c.Disconnect(250)
}
