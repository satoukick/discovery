package model

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNodeStatus(t *testing.T) {
	Convey("test NodeStatus constants", t, func() {
		So(NodeStatusUP, ShouldEqual, 0)
		So(NodeStatusLost, ShouldEqual, 1)
	})
}

func TestAppID(t *testing.T) {
	Convey("test AppID constant", t, func() {
		So(AppID, ShouldEqual, "infra.discovery")
	})
}

func TestNode(t *testing.T) {
	Convey("test Node struct initialization", t, func() {
		node := &Node{
			Addr:   "127.0.0.1:8080",
			Status: NodeStatusUP,
			Zone:   "sh001",
		}

		So(node.Addr, ShouldEqual, "127.0.0.1:8080")
		So(node.Status, ShouldEqual, NodeStatusUP)
		So(node.Zone, ShouldEqual, "sh001")
	})

	Convey("test Node with lost status", t, func() {
		node := &Node{
			Addr:   "192.168.1.1:9090",
			Status: NodeStatusLost,
			Zone:   "bj001",
		}

		So(node.Status, ShouldEqual, NodeStatusLost)
	})
}

func TestScheduler(t *testing.T) {
	Convey("test Scheduler struct initialization", t, func() {
		scheduler := &Scheduler{
			AppID:   "test.app",
			Env:     "production",
			Clients: make(map[string]*ZoneStrategy),
			Remark:  "test scheduler",
		}

		So(scheduler.AppID, ShouldEqual, "test.app")
		So(scheduler.Env, ShouldEqual, "production")
		So(scheduler.Remark, ShouldEqual, "test scheduler")
		So(scheduler.Clients, ShouldNotBeNil)
		So(len(scheduler.Clients), ShouldEqual, 0)
	})

	Convey("test Scheduler.Set with valid JSON", t, func() {
		scheduler := &Scheduler{}
		jsonContent := `{
			"app_id": "test.app.id",
			"env": "staging",
			"clients": {
				"sh001": {
					"zones": {
						"sh002": {
							"weight": 100
						}
					}
				}
			},
			"remark": "test remark"
		}`

		err := scheduler.Set(jsonContent)
		So(err, ShouldBeNil)
		So(scheduler.AppID, ShouldEqual, "test.app.id")
		So(scheduler.Env, ShouldEqual, "staging")
		So(scheduler.Remark, ShouldEqual, "test remark")
		So(scheduler.Clients, ShouldNotBeNil)
		So(len(scheduler.Clients), ShouldEqual, 1)

		zoneStrategy, ok := scheduler.Clients["sh001"]
		So(ok, ShouldBeTrue)
		So(zoneStrategy, ShouldNotBeNil)
		So(len(zoneStrategy.Zones), ShouldEqual, 1)

		strategy, ok := zoneStrategy.Zones["sh002"]
		So(ok, ShouldBeTrue)
		So(strategy.Weight, ShouldEqual, 100)
	})

	Convey("test Scheduler.Set with invalid JSON", t, func() {
		scheduler := &Scheduler{}
		invalidJSON := `{invalid json}`

		err := scheduler.Set(invalidJSON)
		So(err, ShouldNotBeNil)
	})

	Convey("test Scheduler.Set with empty JSON", t, func() {
		scheduler := &Scheduler{}
		emptyJSON := `{}`

		err := scheduler.Set(emptyJSON)
		So(err, ShouldBeNil)
		So(scheduler.AppID, ShouldEqual, "")
		So(scheduler.Env, ShouldEqual, "")
	})
}

func TestZoneStrategy(t *testing.T) {
	Convey("test ZoneStrategy struct", t, func() {
		zoneStrategy := &ZoneStrategy{
			Zones: make(map[string]*Strategy),
		}
		zoneStrategy.Zones["zone1"] = &Strategy{Weight: 50}
		zoneStrategy.Zones["zone2"] = &Strategy{Weight: 50}

		So(len(zoneStrategy.Zones), ShouldEqual, 2)
		So(zoneStrategy.Zones["zone1"].Weight, ShouldEqual, 50)
		So(zoneStrategy.Zones["zone2"].Weight, ShouldEqual, 50)
	})
}

func TestStrategy(t *testing.T) {
	Convey("test Strategy struct with various weights", t, func() {
		strategy1 := &Strategy{Weight: 100}
		strategy2 := &Strategy{Weight: 0}
		strategy3 := &Strategy{Weight: -1}

		So(strategy1.Weight, ShouldEqual, 100)
		So(strategy2.Weight, ShouldEqual, 0)
		So(strategy3.Weight, ShouldEqual, -1)
	})
}

func TestZone(t *testing.T) {
	Convey("test Zone struct", t, func() {
		zone := &Zone{
			Src: "sh001",
			Dst: map[string]int{
				"bj001": 30,
				"sh002": 70,
			},
		}

		So(zone.Src, ShouldEqual, "sh001")
		So(len(zone.Dst), ShouldEqual, 2)
		So(zone.Dst["bj001"], ShouldEqual, 30)
		So(zone.Dst["sh002"], ShouldEqual, 70)
	})

	Convey("test Zone with empty destination", t, func() {
		zone := &Zone{
			Src: "test001",
			Dst: make(map[string]int),
		}

		So(zone.Src, ShouldEqual, "test001")
		So(len(zone.Dst), ShouldEqual, 0)
	})
}

func TestSchedulerComplexScenario(t *testing.T) {
	Convey("test Scheduler with complex multi-zone configuration", t, func() {
		scheduler := &Scheduler{}
		complexJSON := `{
			"app_id": "main.arch.service",
			"env": "production",
			"clients": {
				"sh001": {
					"zones": {
						"sh002": {"weight": 50},
						"bj001": {"weight": 30},
						"gz001": {"weight": 20}
					}
				},
				"bj001": {
					"zones": {
						"bj002": {"weight": 60},
						"sh001": {"weight": 40}
					}
				}
			},
			"remark": "production scheduler config"
		}`

		err := scheduler.Set(complexJSON)
		So(err, ShouldBeNil)
		So(scheduler.AppID, ShouldEqual, "main.arch.service")
		So(len(scheduler.Clients), ShouldEqual, 2)

		// Verify sh001 client zones
		sh001Strategy := scheduler.Clients["sh001"]
		So(sh001Strategy, ShouldNotBeNil)
		So(len(sh001Strategy.Zones), ShouldEqual, 3)
		So(sh001Strategy.Zones["sh002"].Weight, ShouldEqual, 50)
		So(sh001Strategy.Zones["bj001"].Weight, ShouldEqual, 30)
		So(sh001Strategy.Zones["gz001"].Weight, ShouldEqual, 20)

		// Verify bj001 client zones
		bj001Strategy := scheduler.Clients["bj001"]
		So(bj001Strategy, ShouldNotBeNil)
		So(len(bj001Strategy.Zones), ShouldEqual, 2)
		So(bj001Strategy.Zones["bj002"].Weight, ShouldEqual, 60)
		So(bj001Strategy.Zones["sh001"].Weight, ShouldEqual, 40)
	})
}
