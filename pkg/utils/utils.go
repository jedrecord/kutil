/*
Copyright: 2020 Jed Record

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; Version 2 (GPLv2)

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

Full license text at: https://gnu.org/licenses/gpl-2.0.txt
*/

// Package utils Some basic formating and maths for Kubernetes resource quantaties
package utils

import (
	"fmt"
)

// CalcPct Calculate percentage of a resource
func CalcPct(avail int64, inuse int64) int64 {
	var pct int64 = 0
	if avail > 0 {
		pct = 100 * inuse / avail
	}
	return pct
}

// FmtCPU Converts Milli CPU values to fractions of CPU threads
func FmtCPU(i int64) string {
	return fmt.Sprintf("%.1f vCPU", float64(i)/1000)
}

// FmtMilli Append 'm' to for Milli CPU values
func FmtMilli(i int64) string {
	return fmt.Sprintf("%vm", i)
}

// FmtGiB Convert Kibibyte (Ki) values to Gibibyte (GiB)
func FmtGiB(i int64) string {
	return fmt.Sprintf("%.f GiB", float64(i)/1024/1024/1024)
}

// FmtMiB Convert Kibibyte (Ki) values to Mebibyte (MiB)
func FmtMiB(i int64) string {
	return fmt.Sprintf("%.f MiB", float64(i)/1024/1024)
}

// FmtPct Convert an int64 to string percentage
func FmtPct(num int64) string {
	return fmt.Sprintf("%d%%", num)
}

// MaxInt returns the larger of x or y.
func MaxInt(x, y int) int {
	if x < y {
		return y
	}
	return x
}


