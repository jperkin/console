/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package console

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	cmdTcGet = unix.TCGETS
	cmdTcSet = unix.TCSETS
)

type strioctl struct {
	ic_cmd     int
	ic_timeout int
	ic_len     int
	ic_dp      uintptr
}

const (
	ptem   = "ptem"
	ldterm = "ldterm"
)

func unlockpt(f *os.File) error {
	var istr strioctl

	istr.ic_cmd = ('P'<<8 | 2) // UNLKPT
	istr.ic_len = 0
	istr.ic_timeout = 0
	istr.ic_dp = 0

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), unix.I_STR, uintptr(unsafe.Pointer(&istr))); err != 0 {
		return err
	}

	clientname, err := ptsname(f)
	if err != nil {
		return err
	}

	client, err := os.OpenFile(clientname, unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		return err
	}

	b := append([]byte(ptem), 0)
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, client.Fd(), unix.I_PUSH, uintptr(unsafe.Pointer(&b[0]))); err != 0 {
		return err
	}

	b = append([]byte(ldterm), 0)
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, client.Fd(), unix.I_PUSH, uintptr(unsafe.Pointer(&b[0]))); err != 0 {
		return err
	}

	return nil
}

func ptsname(f *os.File) (string, error) {
	var istr strioctl
	var status syscall.Stat_t

	istr.ic_cmd = ('P'<<8 | 1) // ISPTM
	istr.ic_len = 0
	istr.ic_timeout = 0
	istr.ic_dp = 0

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), unix.I_STR, uintptr(unsafe.Pointer(&istr))); err != 0 {
		return "", err
	}

	if err := syscall.Fstat(int(f.Fd()), &status); err != nil {
		return "", err
	}

	return fmt.Sprintf("/dev/pts/%d", unix.Minor(status.Rdev)), nil
}
