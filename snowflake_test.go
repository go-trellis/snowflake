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
	"testing"
)

// Benchmarks Presence Update event with fake data.
func BenchmarkNext(b *testing.B) {

	worker, _ := NewWorker(1, 1)

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		worker.Next()
	}
}

func BenchmarkNextMaxSequence(b *testing.B) {
	SetMaxNode(0, 0, 22)

	worker, _ := NewWorker(0, 0)

	b.ReportAllocs()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		worker.Next()
	}
}

func BenchmarkNextNoSequence(b *testing.B) {
	SetMaxNode(5, 5, 0)

	worker, _ := NewWorker(0, 0)

	b.ReportAllocs()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		worker.Next()
	}
}

func BenchmarkNextSleep(b *testing.B) {

	worker, _ := NewWorker(1, 1)

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		worker.NextSleep()
	}

}

func BenchmarkNextSleepMaxSequence(b *testing.B) {
	SetMaxNode(0, 0, 22)

	worker, _ := NewWorker(0, 0)

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		worker.NextSleep()
	}

}

func BenchmarkNextSleepNoSequence(b *testing.B) {
	SetMaxNode(5, 5, 0)

	worker, _ := NewWorker(0, 0)

	b.ReportAllocs()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		worker.NextSleep()
	}
}
