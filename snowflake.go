/*
Copyright © 2020 Henry Huang <hhh@rutcode.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package snowflake

import (
	"fmt"
	"sync"
	"time"
)

var (
	defEpoch        int64 = 1577808000000
	defTimeAccuracy int64 = 1000000
	nodeBits        int64 = 10
	sequenceBits    int64 = 12
)

// Worker 工作对象
type Worker struct {
	locker sync.Mutex

	conf *Config

	lastTimestamp int64

	sequence int64

	maxNodeID    int64
	sequenceMask int64

	nodeIDShift        int64
	nodeIDShiftResult  int64
	timestampLeftShift int64
}

// NewWorker 生产工作对象
func NewWorker(nodeID int64) (*Worker, error) {

	c := &Config{
		NodeID:       nodeID,
		Eponch:       defEpoch,
		SequenceBits: sequenceBits,
		NodesBits:    nodeBits,
		timeAccuracy: defTimeAccuracy,
	}

	return NewWorkerWithConfig(c)
}

// NewWorkerWithConfig 通过配置文件生成对象
func NewWorkerWithConfig(c *Config) (*Worker, error) {
	if err := c.init(); err != nil {
		return nil, err
	}

	w := &Worker{
		conf:          c,
		lastTimestamp: -1,

		maxNodeID:    -1 ^ (-1 << c.NodesBits),
		sequenceMask: -1 ^ (-1 << sequenceBits),
	}

	if w.conf.NodeID > w.maxNodeID || w.conf.NodeID < 0 {
		return nil, fmt.Errorf("node id can't be greater than %d or less than 0", w.conf.NodeID)
	}
	w.nodeIDShift = w.conf.SequenceBits
	w.nodeIDShiftResult = w.conf.NodeID << w.nodeIDShift
	w.timestampLeftShift = w.conf.SequenceBits + w.conf.NodesBits
	return w, nil
}

// SetDefEpochTime 设置默认开始时间
// epoch 判断长度设置精度，10位为秒，13位为毫秒，16位微妙
// 设置的起始时间不能大于int64的取值范围，即当前时间减去设置时间的毫秒数，不能超过2的63次方
// 这里移位运算为22位，则，设置的时间不能超过 2的63次方向右移22位（代表毫秒数）
// 默认开始时间 2020-01-01 00:00:00.000 UTC
func SetDefEpochTime(epoch int64) error {

	acc, err := checkEponch(epoch)
	if err != nil {
		return err
	}
	timenow := time.Now()

	x := (timenow.UnixNano()/acc - acc) << (sequenceBits + nodeBits)

	if x < 0 {
		timenow = timenow.Add(time.Duration(-9223372036854775808 >> (sequenceBits + nodeBits) * acc))
		return fmt.Errorf("you can't set epoch time before : %+v", timenow)
	}

	defEpoch = epoch
	defTimeAccuracy = acc
	return nil
}

// SetMaxNode 设置最大节点信息
// 默认为10, 12
func SetMaxNode(maxNodeBits, maxSequenceBits int64) {
	if maxNodeBits < 0 || maxSequenceBits < 0 ||
		maxNodeBits+maxSequenceBits > 63 {
		panic(fmt.Errorf("max bits can't be greater than 63 (maxNodeBits+maxSequenceBits) or less than 0"))
	}
	// workerBits = maxDatacenterBits
	// datacenterBits = maxWorkerBits

	sequenceBits = maxSequenceBits
}

// Next 获取下一个ID值
// 统一时刻只能被调用一次
func (p *Worker) Next() (int64, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	timestamp := p.timeGen()
	if timestamp < p.lastTimestamp {
		return 0, fmt.Errorf("Clock moved backwards. Refusing to generate id for %d milliseconds",
			p.lastTimestamp-timestamp)
	}

	if p.lastTimestamp == timestamp {
		p.sequence = (p.sequence + 1) & p.sequenceMask
		if p.sequence == 0 {
			timestamp = p.tilNextTimestamp()
		}
	} else {
		p.sequence = 0
	}

	p.lastTimestamp = timestamp

	return ((timestamp - p.conf.Eponch) << p.timestampLeftShift) | p.nodeIDShiftResult | p.sequence, nil
}

// NextSleep 获取下一个ID值
// 时间戳一样，则沉睡1个单位时间
// 统一时刻只能被调用一次
func (p *Worker) NextSleep() (int64, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	timestamp := p.timeGen()
	if timestamp < p.lastTimestamp {
		return 0, fmt.Errorf("Clock moved backwards. Refusing to generate id for %d milliseconds",
			p.lastTimestamp-timestamp)
	}

	if p.lastTimestamp == timestamp {
		p.sequence = (p.sequence + 1) & p.sequenceMask
		if p.sequence == 0 {
			timestamp = p.timeGen()
			for timestamp <= p.lastTimestamp {
				time.Sleep(time.Duration(p.conf.timeAccuracy))
				timestamp = p.timeGen()
			}
		}
	} else {
		p.sequence = 0
	}

	p.lastTimestamp = timestamp

	return ((timestamp - p.conf.Eponch) << p.timestampLeftShift) | p.nodeIDShiftResult | p.sequence, nil
}

// GetEpochTime 获取开始时间
func (p *Worker) GetEpochTime() int64 {
	return p.conf.Eponch
}

func (p *Worker) tilNextTimestamp() int64 {
	timestamp := p.timeGen()
	for timestamp <= p.lastTimestamp {
		timestamp = p.timeGen()
	}
	return timestamp
}

func (p *Worker) timeGen() int64 {
	return time.Now().UnixNano() / p.conf.timeAccuracy
}
