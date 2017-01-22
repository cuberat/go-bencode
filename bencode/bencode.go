// Copyright (c) 2017 Don Owens <don@regexguy.com>.  All rights reserved.
//
// This software is released under the BSD license:
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
//  * Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
//
//  * Redistributions in binary form must reproduce the above
//    copyright notice, this list of conditions and the following
//    disclaimer in the documentation and/or other materials provided
//    with the distribution.
//
//  * Neither the name of the author nor the names of its
//    contributors may be used to endorse or promote products derived
//    from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS
// FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
// COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
// HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED
// OF THE POSSIBILITY OF SUCH DAMAGE.


package bencode

import (
    "bufio"
    "fmt"
    // "log"
    "io"
    "strconv"
)

type Decoder struct {
    r *bufio.Reader
}

// A Delim is a byte representing the start or end of a list or dictionary:
//     [ ] { }
type Delim byte

// A Token holds a value of one of these types:
//
//     Delim, representing the beginning lists and dictionaries: l d
//         or the end of one: e
//     int64, for integers
//     []byte, for byte strings
type Token interface{}

// func Unmarshal(data []byte, v interface {}) error {

// }

func NewDecoder(r io.Reader) *Decoder {
    dec := new(Decoder)
    dec.r = bufio.NewReader(r)

    return dec
}

func (dec *Decoder) Token() (Token, error) {
    r := dec.r

    b := []byte{'\n'}
    
    _, err := r.Read(b)
    if err != nil {
        return nil, err
    }

    s := b[0]
    switch {
    case s == 'i':
        // integer
        num, err := dec.get_int('e')
        return num, err
    case s == 'l':
        // list
        return Delim('l'), nil
    case s == 'd':
        // dictionary
        return Delim('d'), nil
    case s == 'e':
        return Delim('e'), nil
    case s >= '0' && s <= '9':
        r.UnreadByte()
        return dec.get_string()
    default:
        return nil, fmt.Errorf("unexpected byte %s", s)
    }

    return nil, nil
}

func (dec *Decoder) get_string() ([]byte, error) {
    size_64, err := dec.get_int(':')
    if err != nil {
        return nil, err
    }
    size := int(size_64)

    // log.Printf("get_string(): size=%d", size)

    p := make([]byte, size, size)
    p_read := p[:]
    amtread := 0

    r := dec.r
    for n, err := r.Read(p_read); amtread < size; n, err = r.Read(p_read) {
        // log.Printf("p_read (%d): '%s'", n, p_read)
        amtread += n

        if err != nil {
            if err == io.EOF {
                break
            }
            return nil, err
        }

        p_read = p_read[n:]
    }

    if amtread < size {
        return nil, fmt.Errorf("short read while reading string")
    }

    return p, nil
}

func (dec *Decoder) get_int(end byte) (int64, error) {
    r := dec.r
    b := []byte{'\n'}
    digits := make([]byte, 0, 1)

    for {
        _, err := r.Read(b)
        if err != nil {
            return 0, err
        }

        d := b[0]

        if d >= '0' && d <= '9' {
            digits = append(digits, d)
            continue
        }

        if d == end {
            // done
            break
        }

        return 0, fmt.Errorf("unexpected byte '%s' in integer spec", d)
    }

    return strconv.ParseInt(string(digits), 10, 64)
}

