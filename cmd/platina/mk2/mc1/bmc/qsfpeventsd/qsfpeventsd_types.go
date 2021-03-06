// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qsfpeventsd

var specComplianceValues = [...]string{
	0x00: "Unspecified",
	0x01: "40G Active Cable (XLPPI)",
	0x02: "40GBASE-LR4",
	0x04: "40GBASE-SR4",
	0x08: "40GBASE-CR4",
	0x10: "10GBASE-SR",
	0x20: "10GBASE-LR",
	0x40: "10GBASE_LRM",
	0x80: "Extended",
}

var extSpecComplianceValues = [...]string{
	0x00: "Unspecified",
	0x01: "100G AOC",
	0x02: "100GBASE-SR4",
	0x03: "100GBASE-LR4",
	0x04: "100GBASE-ER4",
	0x05: "100GBASE-SR10",
	0x06: "100G CWDM4",
	0x07: "100G PSRM4 Parallel SMF",
	0x08: "100G ACC",
	0x0B: "100GBASE-CR4",
	0x0C: "25GBASE-CR CA-S",
	0x0D: "25GBASE-CR CA-N",
	0x10: "40GBASE-ER4",
	0x11: "4 x 10GBASE-SR",
	0x12: "40G PSM4 Parallel SMF",
	0x13: "G959.1 profile P1I1-2D1",
	0x14: "G959.1 profile P1S1-2D2",
	0x15: "G959.1 profile P1L1-2D2",
	0x16: "10GBASE-T SFI",
	0x17: "100G CLR4",
	0x18: "100G AOC",
	0x19: "100G ACC",
	0x1A: "100GE-DWDM2",
}

var redisFields = []string{
	"qsfp.alarms",
	"qsfp.compliance",
	"qsfp.partnumber",
	"qsfp.rx.power.highAlarmThreshold.units.mW",
	"qsfp.rx.power.highWarnThreshold.units.mW",
	"qsfp.rx.power.lowAlarmThreshold.units.mW",
	"qsfp.rx.power.lowWarnThreshold.units.mW",
	"qsfp.rx1.alarms",
	"qsfp.rx1.power.units.mW",
	"qsfp.rx2.alarms",
	"qsfp.rx2.power.units.mW",
	"qsfp.rx3.alarms",
	"qsfp.rx3.power.units.mW",
	"qsfp.rx4.alarms",
	"qsfp.rx4.power.units.mW",
	"qsfp.serialnumber",
	"qsfp.temperature.units.C",
	"qsfp.temperature.highAlarmThreshold.units.C",
	"qsfp.temperature.highWarnThreshold.units.C",
	"qsfp.temperature.lowAlarmThreshold.units.C",
	"qsfp.temperature.lowWarnThreshold.units.C",
	"qsfp.tx.biasHighAlarmThreshold.units.mA",
	"qsfp.tx.biasHighWarnThreshold.units.mA",
	"qsfp.tx.biasLowAlarmThreshold.units.mA",
	"qsfp.tx.biasLowWarnThreshold.units.mA",
	"qsfp.tx.power.highAlarmThreshold.units.mW",
	"qsfp.tx.power.highWarnThreshold.units.mW",
	"qsfp.tx.power.lowAlarmThreshold.units.mW",
	"qsfp.tx.power.lowWarnThreshold.units.mW",
	"qsfp.tx1.alarms",
	"qsfp.tx1.bias.units.mA",
	"qsfp.tx1.power.units.mW",
	"qsfp.tx2.alarms",
	"qsfp.tx2.bias.units.mA",
	"qsfp.tx2.power.units.mW",
	"qsfp.tx3.alarms",
	"qsfp.tx3.bias.units.mA",
	"qsfp.tx3.power.units.mW",
	"qsfp.tx4.alarms",
	"qsfp.tx4.bias.units.mA",
	"qsfp.tx4.power.units.mW",
	"qsfp.vcc.units.V",
	"qsfp.vcc.highAlarmThreshold.units.V",
	"qsfp.vcc.highWarnThreshold.units.V",
	"qsfp.vcc.lowAlarmThreshold.units.V",
	"qsfp.vcc.lowWarnThreshold.units.V",
	"qsfp.vendor",
}
