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
	"os"
)

// LogError Error logging
func LogError(msg string) {
	fmt.Printf("Error: %s\n", msg)
	os.Exit(1)
}

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
	// Storing precision in var p
	var p int = 0
	// If non-fraction, use precision 0. Otherwise show 1 decimal point precision
	r := fmt.Sprintf("%.1f", float64(i)/1000)
	if r[len(r)-1:] != "0" {
		p = 1
	}
	return fmt.Sprintf("%.*f vCPU", p, float64(i)/1000)
}

// FmtMilli Append 'm' to for Milli CPU values
func FmtMilli(i int64) string {
	return fmt.Sprintf("%vm", i)
}

// FmtMem Convert Kibibyte (Ki) values to Gibibyte (GiB) or Mebibyte (MiB)
func FmtMem(i int64) string {
	// Storing precision in var p
	var p int = 0
	// If mem is an even GiB, use precision 0. Otherwise show 1 decimal point precision
	r := fmt.Sprintf("%.1f", float64(i)/1024/1024/1024)
	if r[len(r)-1:] != "0" {
		p = 1
	}
	if i/1024/1024/1024 > 0 {
		return fmt.Sprintf("%.*f GiB", p, float64(i)/1024/1024/1024)
	}
	return fmt.Sprintf("%.f MiB", float64(i)/1024/1024)
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


