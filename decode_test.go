// Copyright 2019 Filip Kroča. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teltonikaparser

import (
	"encoding/binary"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecodeCodec8(t *testing.T) {
	const packet = "007ccafe0133000f33353230393430383136373231373908020000016c32b488a0000a7a367c1d30018700000000000000f1070301001500ef000342318bcd42dcce606401f1000059d9000000016c32b48c88000a7a367c1d3001870000000000000015070301001501ef0003423195cd42dcce606401f1000059d90002"

	bs, err := hex.DecodeString(packet)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := Decode(&bs)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if decoded.IMEI != "352094081672179" {
		t.Fatalf("IMEI: got %q", decoded.IMEI)
	}
	if decoded.CodecID != 0x08 {
		t.Fatalf("CodecID: got %#x", decoded.CodecID)
	}
	if decoded.NoOfData != 2 || len(decoded.Data) != 2 {
		t.Fatalf("NoOfData: got %d records %d", decoded.NoOfData, len(decoded.Data))
	}
	if decoded.Data[0].EventID != 241 {
		t.Fatalf("EventID: got %d", decoded.Data[0].EventID)
	}
	wantResp := []byte{0x00, 0x05, 0xca, 0xfe, 0x01, 0x33, 0x02}
	if string(decoded.Response) != string(wantResp) {
		t.Fatalf("Response: got %x want %x", decoded.Response, wantResp)
	}
}

func TestDecodeCodec8Extended(t *testing.T) {
	const packet = "0086cafe0101000f3335323039333038353639383230368e0100000167efa919800200000000000000000000000000000000fc0013000800ef0000f00000150500c80000450200010000710000fc00000900b5000000b600000042305600cd432a00ce6064001100090012ff22001303d1000f0000000200f1000059d90010000000000000000001"

	bs, err := hex.DecodeString(packet)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := Decode(&bs)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if decoded.CodecID != 0x8e {
		t.Fatalf("CodecID: got %#x", decoded.CodecID)
	}
	if decoded.NoOfData != 1 {
		t.Fatalf("NoOfData: got %d", decoded.NoOfData)
	}
	if decoded.Data[0].Lat != 0 || decoded.Data[0].Lng != 0 {
		t.Fatalf("expected GPS 0,0 without fix, got lat=%d lng=%d", decoded.Data[0].Lat, decoded.Data[0].Lng)
	}
}

func TestDecodeRejectsTooShort(t *testing.T) {
	bs := []byte{0xff}
	if _, err := Decode(&bs); err == nil {
		t.Fatal("expected error")
	}
}

func TestDecodeRejectsNonTeltonika(t *testing.T) {
	bs := make([]byte, 45)
	if _, err := Decode(&bs); err == nil {
		t.Fatal("expected error")
	}
}

func TestCutIOxLenDoesNotPanicOnTruncatedValue(t *testing.T) {
	bs := make([]byte, 6)
	binary.BigEndian.PutUint16(bs[0:], 16)
	binary.BigEndian.PutUint16(bs[2:], 1000)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()

	_, err := cutIOxLen(&bs, 0)
	if err == nil {
		t.Fatal("expected bounds error")
	}
}

func TestDecodeRegressionFixtures(t *testing.T) {
	fixtures := []string{
		"issue-6-fm3001.hex",
		"issue-7-fmb003.hex",
	}

	for _, name := range fixtures {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			hexStr := strings.TrimSpace(string(raw))
			bs, err := hex.DecodeString(hexStr)
			if err != nil {
				t.Fatalf("hex: %v", err)
			}

			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic: %v", r)
				}
			}()

			decoded, err := Decode(&bs)
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if decoded.NoOfData == 0 || len(decoded.Data) == 0 {
				t.Fatal("expected AVL records")
			}
		})
	}
}

func TestHumanUnknownElementReturnsError(t *testing.T) {
	h := HumanDecoder{}
	el := Element{Length: 1, IOID: 65535, Value: []byte{0x01}}
	if _, err := h.Human(&el, "FMBXY"); err == nil {
		t.Fatal("expected error for unknown IO ID")
	}
}

func TestCRC16IBMGetinfo(t *testing.T) {
	// Codec 12 payload for command "getinfo" without preamble and CRC.
	payload, err := hex.DecodeString("0c010500000007676574696e666f01")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := crc16IBM(payload), uint16(0x4312); got != want {
		t.Fatalf("crc16IBM = %#04x, want %#04x", got, want)
	}
}
