// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stringutils2

func IsUtf8(str string) bool {
	for _, runeVal := range str {
		if runeVal > 0x7f {
			return true
		}
	}
	return false
}

func IsPrintableAscii(b byte) bool {
	if b >= 32 && b <= 126 {
		return true
	}
	return false
}

func IsPrintableAsciiString(str string) bool {
	for _, b := range []byte(str) {
		if !IsPrintableAscii(b) {
			return false
		}
	}
	return true
}
