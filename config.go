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
)

// Config 配置
type Config struct {
	NodeID       int64
	Eponch       int64 // Def 2020-01-01 00:00:00.000
	SequenceBits int64 //
	NodesBits    int64 //
	// TimeBits     int64 //
	timeAccuracy int64 // Millsecond = 1000000
}

func (p *Config) init() error {
	if p == nil {
		return fmt.Errorf("config should not be nil")
	}
	if p.Eponch == 0 {
		p.Eponch = defEpoch
		p.timeAccuracy = defTimeAccuracy
	} else {
		acc, err := checkEponch(p.Eponch)
		if err != nil {
			return err
		}
		p.timeAccuracy = acc
	}

	if p.SequenceBits < 0 || p.NodesBits < 0 {
		return fmt.Errorf("bits can't less than 0")
	}

	if (p.SequenceBits + p.NodesBits) > 63 {
		return fmt.Errorf("sum of bits can't greater than 63")
	}

	return nil
}

func checkEponch(eponch int64) (int64, error) {

	twepochLen := len(fmt.Sprintf("%d", eponch))
	var timeAccuracy int64 = 1000000
	switch twepochLen {
	case 10: //秒
		timeAccuracy = 1000000 * 1000
	case 13: //毫秒
		timeAccuracy = 1000000
	case 16: //微秒
		timeAccuracy = 1000
	default:
		return 0, fmt.Errorf("eponch's length should be 10, 13 or 16")
	}
	return timeAccuracy, nil
}
