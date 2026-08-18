# teltonikaparser

[![Go Reference](https://pkg.go.dev/badge/github.com/filipkroca/teltonikaparser.svg)](https://pkg.go.dev/github.com/filipkroca/teltonikaparser)
[![CI](https://github.com/filipkroca/teltonikaparser/actions/workflows/ci.yml/badge.svg)](https://github.com/filipkroca/teltonikaparser/actions/workflows/ci.yml)

Go parser for [Teltonika](https://wiki.teltonika-gps.com/view/Codec) AVL packets:

- **Codec 8** and **Codec 8 Extended** over **UDP** (`Decode`)
- **Codec 12** text commands (`EncodeCommandRequest`, `DecodeCommandRequest`, `DecodeCommandResponse`)
- Optional human-readable IO names for device families FMBXY, FM64, FM36, and FM11XY

`Decode` is a low-level implementation. On a typical core it runs at roughly **1M+ packets/s** (`Decode` ~788 ns/op).

## Install

```bash
go get github.com/filipkroca/teltonikaparser
```

Requires **Go 1.22+**.

## UDP vs TCP

`Decode` currently accepts **UDP** packets only. Those start with a 2-byte length and packet ID `0xCAFE`.

Teltonika **TCP** uses a different framing: preamble `0x00000000`, 4-byte data length, AVL payload, CRC-16/IBM. TCP is not implemented yet. See [issue #11](https://github.com/filipkroca/teltonikaparser/issues/11).

Full codec field layout: [Teltonika wiki](https://wiki.teltonika-gps.com/view/Codec).

## Decode a UDP packet

```go
package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/filipkroca/teltonikaparser"
)

func main() {
	raw, err := hex.DecodeString("007ccafe0133000f33353230393430383136373231373908020000016c32b488a0000a7a367c1d30018700000000000000f1070301001500ef000342318bcd42dcce606401f1000059d9000000016c32b48c88000a7a367c1d3001870000000000000015070301001501ef0003423195cd42dcce606401f1000059d90002")
	if err != nil {
		log.Fatal(err)
	}

	decoded, err := teltonikaparser.Decode(&raw)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("IMEI %s codec %#x records %d\n", decoded.IMEI, decoded.CodecID, decoded.NoOfData)
	// decoded.Response is the UDP ACK to send back to the device
}
```

`Decoded.Data` is a slice of AVL records. Each record has GPS fields and `Elements` (`IOID`, `Length`, raw `Value`).

## Human-readable IO

```go
human := teltonikaparser.HumanDecoder{}
for _, rec := range decoded.Data {
	for i := range rec.Elements {
		h, err := human.Human(&rec.Elements[i], "FMBXY") // or "FM64", "FM36", "FM11XY"
		if err != nil {
			continue // unknown IO ID for this family
		}
		val, err := h.GetFinalValue()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s = %v\n", h.AvlEncodeKey.PropertyName, val)
	}
}
```

## Codec 12 commands

Text commands use the same strings as SMS (`getinfo`, `getio`, …). See `EncodeCommandRequest` and `DecodeCommandRequest`. Spec: [Codec 12](https://wiki.teltonika-gps.com/view/Codec#Codec_12).

## Performance

```
Decode()  788 ns/op   592 B/op   4 allocs/op
Human()  4082 ns/op  4722 B/op  49 allocs/op
```

A 58M real-world UDP packet run on Intel Core i7-7700K finished in 31s (~1.8M packets/s). That dataset is not published (it contains production IMEI numbers). Use `go test -bench=.` for a repeatable micro-benchmark.

## License

BSD-3-Clause. See [LICENSE](LICENSE).
