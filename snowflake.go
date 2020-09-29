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
	twepoch          int64 = 1577808000000
	workerIDBits     int64 = 5
	datacenterIDBits int64 = 5
	sequenceBits     int64 = 12
	// maxWorkerID        int64 = -1 ^ (-1 << workerIDBits)
	// maxDatacenterID    int64 = -1 ^ (-1 << datacenterIDBits)
	// sequenceMask       int64 = -1 ^ (-1 << sequenceBits)
	// workerIDShift      int64 = sequenceBits
	// datacenterIDShift  int64 = sequenceBits + workerIDShift
	// timestampLeftShift int64 = sequenceBits + workerIDBits + datacenterIDBits
)

// Worker 工作对象
type Worker struct {
	locker sync.Mutex

	lastTimestamp int64
	epoch         int64

	workerID     int64
	datacenterID int64
	sequence     int64

	maxWorkerID     int64
	maxDatacenterID int64
	sequenceMask    int64

	workerIDShift      int64
	datacenterIDShift  int64
	timestampLeftShift int64
}

// NewWorker 生产工作对象
func NewWorker(datacenterID, workerID int64) (*Worker, error) {
	workerIDBits, datacenterIDBits, sequenceBits := workerIDBits, datacenterIDBits, sequenceBits

	var maxWorkerID int64 = -1 ^ (-1 << workerIDBits)
	var maxDatacenterID int64 = -1 ^ (-1 << datacenterIDBits)

	if workerID > maxWorkerID || workerID < 0 {
		return nil, fmt.Errorf("worker Id can't be greater than %d or less than 0", workerID)
	}
	if datacenterID > maxDatacenterID || datacenterID < 0 {
		return nil, fmt.Errorf("datacenter Id can't be greater than %d or less than 0", datacenterID)
	}
	w := &Worker{
		workerID:        workerID,
		datacenterID:    datacenterID,
		lastTimestamp:   -1,
		epoch:           twepoch,
		maxWorkerID:     maxWorkerID,
		maxDatacenterID: maxDatacenterID,
		sequenceMask:    -1 ^ (-1 << sequenceBits),
	}
	w.workerIDShift = sequenceBits
	w.datacenterIDShift = sequenceBits + workerIDBits
	w.timestampLeftShift = sequenceBits + workerIDBits + datacenterIDBits
	return w, nil
}

// SetEpochTime 设置默认开始时间
// 设置的起始时间不能大于int64的取值范围，即当前时间减去设置时间的毫秒数，不能超过2的63次方
// 这里移位运算为22位，则，设置的时间不能超过 2的63次方向右移22位（代表毫秒数）
// 默认开始时间 2020-01-01 00:00:00.000 UTC
func SetEpochTime(t time.Time) {
	timenow := time.Now()
	x := (timenow.UnixNano()/1000000 - t.UnixNano()/1000000) << (sequenceBits + workerIDBits + datacenterIDBits)

	if x < 0 {
		timenow = timenow.Add(time.Duration(-9223372036854775808>>22) * time.Millisecond)
		panic(fmt.Errorf("you can't set epoch time before : %+v", timenow))
	}
	twepoch = t.UnixNano() / 1000000
}

// SetMaxNode 设置最大节点信息
// 默认为5，5
func SetMaxNode(maxDatacenterBits, maxWorkerBits, maxSequenceBits int64) {
	if maxDatacenterBits < 0 || maxWorkerBits < 0 || maxSequenceBits < 0 ||
		maxDatacenterBits+maxWorkerBits+maxSequenceBits > 22 {
		panic(fmt.Errorf("max bits can't be greater than 22 (maxDatacenterBits+maxWorkerBits+maxSequenceBits) or less than 0"))
	}
	workerIDBits = maxDatacenterBits
	datacenterIDBits = maxWorkerBits
	sequenceBits = maxSequenceBits
}

// Next 获取下一个ID值
// 统一时刻只能被调用一次
func (p *Worker) Next() (int64, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	timestamp := timeGen()
	if timestamp < p.lastTimestamp {
		return 0, fmt.Errorf("Clock moved backwards. Refusing to generate id for %d milliseconds",
			p.lastTimestamp-timestamp)
	}

	if p.lastTimestamp == timestamp {
		p.sequence = (p.sequence + 1) & p.sequenceMask
		if p.sequence == 0 {
			timestamp = p.tilNextMillis()
		}
	} else {
		p.sequence = 0
	}

	p.lastTimestamp = timestamp

	return ((timestamp - p.epoch) << p.timestampLeftShift) |
			(p.datacenterID << p.datacenterIDShift) |
			(p.workerID << p.workerIDShift) |
			p.sequence,
		nil
}

// NextSleep 获取下一个ID值
// 时间戳一样，则沉睡1毫秒
// 统一时刻只能被调用一次
func (p *Worker) NextSleep() (int64, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	timestamp := timeGen()
	if timestamp < p.lastTimestamp {
		return 0, fmt.Errorf("Clock moved backwards. Refusing to generate id for %d milliseconds",
			p.lastTimestamp-timestamp)
	}

	if p.lastTimestamp == timestamp {
		p.sequence = (p.sequence + 1) & p.sequenceMask
		if p.sequence == 0 {
			timestamp = timeGen()
			for timestamp <= p.lastTimestamp {
				time.Sleep(time.Millisecond)
				timestamp = timeGen()
			}
		}
	} else {
		p.sequence = 0
	}

	p.lastTimestamp = timestamp

	return ((timestamp - p.epoch) << p.timestampLeftShift) |
			(p.datacenterID << p.datacenterIDShift) |
			(p.workerID << p.workerIDShift) |
			p.sequence,
		nil
}

// GetEpochTime 获取开始时间
func (p *Worker) GetEpochTime() int64 {
	return p.epoch
}

func (p *Worker) tilNextMillis() int64 {
	timestamp := timeGen()
	for timestamp <= p.lastTimestamp {
		timestamp = timeGen()
	}
	return timestamp
}

func timeGen() int64 {
	return time.Now().UnixNano() / 1000000
}
