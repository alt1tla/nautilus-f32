---
title: Publishing to MQTT (Sparkplug B)
description: Expose a controller's tags to any Sparkplug-aware SCADA host — faithful types, report-by-exception, and TCK-conformant birth/death.
---

Expose a controller's tags to a Sparkplug host (Ignition, any Sparkplug-aware
SCADA) with the `sparkplug` package. The runtime is the edge node; each
`io.Driver` becomes a device whose birth/death follows its link:

```go
node, _ := sparkplug.New(rt, sparkplug.Config{
    BrokerURL: "tcp://broker:1883", GroupID: "Plant", EdgeNode: "Line1",
    BdSeqFile: "/var/lib/nautilus/line1.bdseq",
},
    // Publish classes are report-by-exception groups, like scan classes:
    sparkplug.WithDefaultRBE(sparkplug.RBE{Deadband: 0.5, MaxInterval: 30 * time.Second}),
    sparkplug.WithPublishClass("fast", sparkplug.RBE{Deadband: 0.1, MaxInterval: 5 * time.Second}),
    sparkplug.WithMetricClass("fast", "PIT_*"),
    // The EtherNet/IP driver's tags become a device; DBIRTH/DDEATH track its health.
    sparkplug.WithDevice(sparkplug.Device{
        ID: "plc1", Tags: driver.InputNames(),
        Health: func() bool { return driver.Health().Connected },
    }),
)
node.Start(ctx)
```

Types map faithfully (BOOL→Boolean, integer→Int64, REAL→Double,
UDT→Template), a SCADA host can write tags back via NCMD, and
`Node Control/Rebirth` is honored.

## Conformance

The node passes the **Sparkplug TCK edge-node profile** — CI runs the
`joyautomation/sparkplug-tck-go` harness against a live node on every push.
MQTT and protobuf live only in this package; the runtime core stays
stdlib-only.
