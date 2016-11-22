/*
   Copyright (c) 2014, Mike 'Fuzzy' Partin <fuzzy@fumanchu.org>
   All rights reserved.

  Redistribution and use in source and binary forms, with or without
  modification, are permitted provided that the following conditions are met:

  1. Redistributions of source code must retain the above copyright notice, this
     list of conditions and the following disclaimer.

  2. Redistributions in binary form must reproduce the above copyright notice,
     this list of conditions and the following disclaimer in the documentation
     and/or other materials provided with the distribution.

  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
  ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
  WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
  DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
  FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
  DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
  SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
  CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
  OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
  OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package gout

import "fmt"

type String string

func (i String) Black() String {
	return String(fmt.Sprintf("\033[30m%s\033[0m", i))
}

func (i String) Red() String {
	return String(fmt.Sprintf("\033[31m%s\033[0m", i))
}

func (i String) Green() String {
	return String(fmt.Sprintf("\033[32m%s\033[0m", i))
}

func (i String) Yellow() String {
	return String(fmt.Sprintf("\033[33m%s\033[0m", i))
}

func (i String) Blue() String {
	return String(fmt.Sprintf("\033[34m%s\033[0m", i))
}

func (i String) Purple() String {
	return String(fmt.Sprintf("\033[35m%s\033[0m", i))
}

func (i String) Cyan() String {
	return String(fmt.Sprintf("\033[36m%s\033[0m", i))
}

func (i String) White() String {
	return String(fmt.Sprintf("\033[37m%s\033[0m", i))
}

func (i String) Bold() String {
	return String(fmt.Sprintf("\033[1m%s\033[0m", i))
}

func (i String) Underline() String {
	return String(fmt.Sprintf("\033[4m%s\033[0m", i))
}

func (i String) Blink() String {
	return String(fmt.Sprintf("\033[5m%s\033[0m", i))
}

func (i String) Reverse() String {
	return String(fmt.Sprintf("\033[7m%s\033[0m", i))
}

func (i String) Conceal() String {
	return String(fmt.Sprintf("\033[8m%s\033[0m", i))
}
