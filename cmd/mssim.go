package main

import (
	"mssim/config"
	"mssim/internal/scenarios/templates"
	"mssim/monitoring"

	"os"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const version = "0.0.1"

var _ueFlags, _multiUesFlags, _amfLoadFlags []cli.Flag

func init() {

	spew.Config.Indent = "  "

	_ueFlags = []cli.Flag{
		&cli.BoolFlag{Name: "disableTunnel", Aliases: []string{"t"}, Usage: "Disable the creation of the GTP-U tunnel interface."},
		&cli.PathFlag{Name: "pcap", Usage: "Capture traffic to given PCAP file when a path is given", Value: "./dump.pcap"},
	}

	_multiUesFlags = []cli.Flag{
		&cli.IntFlag{Name: "number-of-ues", Value: 1, Aliases: []string{"n"}},
		&cli.IntFlag{Name: "timeBetweenRegistration", Value: 500, Aliases: []string{"tr"}, Usage: "The time in ms, between UE registration."},
		&cli.IntFlag{Name: "timeBeforeDeregistration", Value: 0, Aliases: []string{"td"}, Usage: "The time in ms, before a UE deregisters once it has been registered. 0 to disable auto-deregistration."},
		&cli.IntFlag{Name: "timeBeforeNgapHandover", Value: 0, Aliases: []string{"ngh"}, Usage: "The time in ms, before triggering a UE handover using NGAP Handover. 0 to disable handover. This requires at least two gNodeB, eg: two N2/N3 IPs."},
		&cli.IntFlag{Name: "timeBeforeXnHandover", Value: 0, Aliases: []string{"xnh"}, Usage: "The time in ms, before triggering a UE handover using Xn Handover. 0 to disable handover. This requires at least two gNodeB, eg: two N2/N3 IPs."},
		&cli.IntFlag{Name: "timeBeforeIdle", Value: 0, Aliases: []string{"idl"}, Usage: "The time in ms, before switching UE to Idle. 0 to disable Idling."},
		&cli.IntFlag{Name: "timeBeforeReconnecting", Value: 1000, Aliases: []string{"tbr"}, Usage: "The time in ms, before reconnecting to gNodeB after switching to Idle state. Default is 1000 ms. Only work in conjunction with timeBeforeIdle."},
		&cli.IntFlag{Name: "numPduSessions", Value: 1, Aliases: []string{"nPdu"}, Usage: "The number of PDU Sessions to create"},
		&cli.BoolFlag{Name: "loop", Aliases: []string{"l"}, Usage: "Register UEs in a loop."},
		&cli.BoolFlag{Name: "tunnel", Aliases: []string{"t"}, Usage: "Enable the creation of the GTP-U tunnel interface."},
		&cli.BoolFlag{Name: "tunnel-vrf", Value: true, Usage: "Enable/disable VRP usage of the GTP-U tunnel interface."},
		&cli.BoolFlag{Name: "dedicatedGnb", Aliases: []string{"d"}, Usage: "Enable the creation of a dedicated gNB per UE. Require one IP on N2/N3 per gNB."},
		&cli.PathFlag{Name: "pcap", Usage: "Capture traffic to given PCAP file when a path is given", Value: "./dump.pcap"},
		&cli.PathFlag{Name: "log", Aliases: []string{"lg"}, Usage: "Log file"},
	}
	_amfLoadFlags = []cli.Flag{
		&cli.IntFlag{Name: "number-of-requests", Value: 1, Aliases: []string{"n"}},
		&cli.IntFlag{Name: "time", Value: 1, Aliases: []string{"t"}},
	}
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.PathFlag{Name: "config", Usage: "config file (./config/config.yml)"},
		},
		Commands: []*cli.Command{
			{
				Name:   "ue",
				Usage:  "one UE attached to one gnB\n",
				Flags:  _ueFlags,
				Action: runUeCmd,
			},
			{
				Name:   "gnb",
				Usage:  "Launch only a gNB",
				Action: runGnbCmd,
			},
			{
				Name:    "multi-ue-pdu",
				Aliases: []string{"multi-ue"},
				Usage: "\nLoad endurance stress tests.\n" +
					"Example for testing multiple UEs: multi-ue -n 5 \n" +
					"This test case will launch N UEs. See mssim multi-ue --help\n",
				Flags:  _multiUesFlags,
				Action: runMultiUesCmd,
			},
			{
				Name: "amf-load-loop",
				Usage: "\nTest AMF responses in interval\n" +
					"Example for generating 20 requests to AMF per second in interval of 20 seconds: amf-load-loop -n 20 -t 20\n",
				Flags:  _amfLoadFlags,
				Action: runAmfLoadTestCmd,
			},
			{
				Name: "amf-availability",
				Usage: "\nTest availability of AMF in interval\n" +
					"Test availability of AMF in 20 seconds: amf-availability -t 20\n",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "time", Value: 1, Aliases: []string{"t"}},
				},
				Action: runAmfTestCmd,
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runUeCmd(c *cli.Context) error {
	name := "Testing an ue attached with configuration"
	cfg := setConfig(*c)
	tunnelEnabled := !c.Bool("disableTunnel")

	log.Info("MSsim version " + version)
	log.Info("---------------------------------------")
	log.Info("[TESTER] Starting test function: ", name)
	log.Info("[TESTER][UE] Number of UEs: ", 1)
	log.Info("[TESTER][UE] disableTunnel is ", !tunnelEnabled)
	log.Info("[TESTER][GNB] Control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port, "~")
	log.Info("[TESTER][GNB] Data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
	for _, amf := range cfg.AMFs {
		log.Info("[TESTER][AMF] AMF IP/Port: ", amf.Ip, "/", amf.Port)
	}
	log.Info("---------------------------------------")

	if c.IsSet("pcap") {
		monitoring.CaptureTraffic(c.Path("pcap"))
	}

	templates.TestAttachUeWithConfiguration(tunnelEnabled)
	return nil
}
func runGnbCmd(c *cli.Context) error {
	name := "Test one gnB"
	cfg := setConfig(*c)

	log.Info("MSsim version " + version)
	log.Info("---------------------------------------")
	log.Info("[TESTER] Starting test function: ", name)
	log.Info("[TESTER][GNB] Number of GNBs: ", 1)
	log.Info("[TESTER][GNB] Control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port, "~")
	log.Info("[TESTER][GNB] Data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
	for _, amf := range cfg.AMFs {
		log.Info("[TESTER][AMF] AMF IP/Port: ", amf.Ip, "/", amf.Port)
	}
	log.Info("---------------------------------------")
	templates.TestAttachGnbWithConfiguration()
	return nil
}
func runMultiUesCmd(c *cli.Context) error {
	var numUes int
	name := "Testing registration of multiple UEs"
	cfg := setConfig(*c)
	if c.IsSet("number-of-ues") {
		numUes = c.Int("number-of-ues")
	} else {
		log.Info(c.Command.Usage)
		return nil
	}

	log.Info("MSsim version " + version)
	log.Info("---------------------------------------")
	log.Info("[TESTER] Starting test function: ", name)
	log.Info("[TESTER][UE] Number of UEs: ", numUes)
	log.Info("[TESTER][GNB] gNodeB control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port, "~")
	log.Info("[TESTER][GNB] gNodeB data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
	for _, amf := range cfg.AMFs {
		log.Info("[TESTER][AMF] AMF IP/Port: ", amf.Ip, "/", amf.Port)
	}
	log.Info("---------------------------------------")

	if c.IsSet("pcap") {
		monitoring.CaptureTraffic(c.Path("pcap"))
	}

	tunnelMode := config.TunnelDisabled
	if c.Bool("tunnel") {
		if c.Bool("tunnel-vrf") {
			tunnelMode = config.TunnelVrf
		} else {
			tunnelMode = config.TunnelTun
		}
	}
	templates.TestMultiUesInQueue(numUes, tunnelMode, c.Bool("dedicatedGnb"), c.Bool("loop"), c.Int("timeBetweenRegistration"), c.Int("timeBeforeDeregistration"), c.Int("timeBeforeNgapHandover"), c.Int("timeBeforeXnHandover"), c.Int("timeBeforeIdle"), c.Int("timeBeforeReconnecting"), c.Int("numPduSessions"), c.String("log"))

	return nil
}

func runAmfLoadTestCmd(c *cli.Context) error {
	var time int
	var numRqs int

	name := "Test AMF responses in interval"
	cfg := setConfig(*c)

	numRqs = c.Int("number-of-requests")
	time = c.Int("time")

	log.Info("MSsim version " + version)
	log.Info("---------------------------------------")
	log.Warn("[TESTER] Starting test function: ", name)
	log.Warn("[TESTER][UE] Number of Requests per second: ", numRqs)
	log.Info("[TESTER][GNB] gNodeB control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
	log.Info("[TESTER][GNB] gNodeB data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
	for _, amf := range cfg.AMFs {
		log.Info("[TESTER][AMF] AMF IP/Port: ", amf.Ip, "/", amf.Port)
	}
	log.Info("---------------------------------------")
	log.Warn("[TESTER][GNB] Total of AMF Responses in the interval:", templates.TestRqsLoop(numRqs, time))
	return nil
}

func runAmfTestCmd(c *cli.Context) error {
	var time int

	name := "Test availability of AMF"
	cfg := setConfig(*c)
	time = c.Int("time")

	log.Info("MSsim version " + version)
	log.Info("---------------------------------------")
	log.Warn("[TESTER] Starting test function: ", name)
	log.Warn("[TESTER][UE] Interval of test: ", time, " seconds")
	log.Info("[TESTER][GNB] Control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
	log.Info("[TESTER][GNB] Data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
	for _, amf := range cfg.AMFs {
		log.Info("[TESTER][AMF] AMF IP/Port: ", amf.Ip, "/", amf.Port)
	}
	log.Info("---------------------------------------")
	templates.TestAvailability(time)
	return nil
}

func setConfig(c cli.Context) config.Config {
	var cfg config.Config
	if c.IsSet("config") {
		cfg = config.Load(c.Path("config"))
	} else {
		cfg = config.LoadDefaultConfig()
	}
	return cfg
}
