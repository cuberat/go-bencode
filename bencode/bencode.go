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
    // "bytes"
    "fmt"
    // "log"
    "io"
    "reflect"
    "sort"
    "strconv"
    "strings"
)

type Decoder struct {
    // r *bufio.Reader
    r *breader
}

type Encoder struct {
    w io.Writer
}

// A Delim is a byte representing the start or end of a list or dictionary:
//     [ ] { }
type Delim byte

// A Token holds a value of one of these types:
//
//     Delim, representing the beginning lists and dictionaries: l d
//         or the end of one: e
//     int64, for integers
//     string, for strings
type Token interface{}

// func Unmarshal(data []byte, v interface {}) error {

// }

type breader struct {
    r *bufio.Reader
    pos uint64
}

func ParseString(s string) (interface{}, error) {
    r := strings.NewReader(s)
    return Parse(r)
}

func DecodeString(s string) (interface{}, error) {
    return ParseString(s)
}

func Decode(r io.Reader) (interface{}, error) {
    return Parse(r)
}

// Not completed yet
func Encode(w io.Writer, v interface{}) (error) {
    enc := NewEncoder(w)
    return enc.Encode(v)
}

func Parse(r io.Reader) (interface{}, error) {
    dec := NewDecoder(r)
    v, err := dec.Parse()

    if err == io.EOF {
        err = nil
    }

    return v, err
}

func (r *breader) Read(p []byte) (n int, err error) {
    n, err = r.r.Read(p)
    r.pos += uint64(n)

    return n, err
}

func (r *breader) UnreadByte() error {
    err := r.r.UnreadByte()
    if err == nil {
        r.pos -= 1
    }

    return err
}

func (r *breader) Tell() uint64 {
    return r.pos
}

func new_reader (r io.Reader) (*breader) {
    reader := new(breader)
    reader.r = bufio.NewReader(r)

    return reader
}

func NewEncoder(w io.Writer) *Encoder {
    enc := new(Encoder)
    enc.w = w

    return enc
}

func NewDecoder(r io.Reader) *Decoder {
    dec := new(Decoder)
    dec.r = new_reader(r)
    // dec.r = bufio.NewReader(r)

    return dec
}

// Not completed yet
func (enc *Encoder) Encode(v interface{}) (error) {
    this_type := reflect.TypeOf(v)
    this_kind := this_type.Kind()

    // log.Printf("got kind %s for '%+v'", this_kind.String(), v)

    switch this_kind {
    case reflect.Interface:
        ival := reflect.ValueOf(v).Elem()
        return enc.Encode(ival)
    case reflect.Int:
        fmt.Fprintf(enc.w, "i%de", v.(int))
    case reflect.Int8:
        fmt.Fprintf(enc.w, "i%de", v.(int8))
    case reflect.Int16:
        fmt.Fprintf(enc.w, "i%de", v.(int16))
    case reflect.Int32:
        fmt.Fprintf(enc.w, "i%de", v.(int32))
    case reflect.Int64:
        fmt.Fprintf(enc.w, "i%de", v.(int64))
    case reflect.Uint:
        fmt.Fprintf(enc.w, "i%de", v.(uint))
    case reflect.Uint8:
        fmt.Fprintf(enc.w, "i%de", v.(uint8))
    case reflect.Uint16:
        fmt.Fprintf(enc.w, "i%de", v.(uint16))
    case reflect.Uint32:
        fmt.Fprintf(enc.w, "i%de", v.(uint32))
    case reflect.Uint64:
        fmt.Fprintf(enc.w, "i%de", v.(uint64))

    case reflect.Float32:
        f32 := fmt.Sprintf("%f", v.(float32))
        if err := enc.Encode(f32); err != nil {
            return err
        }

    case reflect.Float64:
        f64 := fmt.Sprintf("%f", v.(float64))
        if err := enc.Encode(f64); err != nil {
            return err
        }

    case reflect.Map:
        return enc.encode_map(v)

    case reflect.Struct:
        return enc.encode_struct(v)

    case reflect.Slice:
        return enc.encode_slice(v)

    case reflect.String:
        s := v.(string)
        fmt.Fprintf(enc.w, "%d:%s", len(s), s)

    case reflect.Array:
        return enc.encode_array(v)

    case reflect.Ptr:
        elem := reflect.ValueOf(v).Elem()
        if ! elem.IsValid() {
            return enc.Encode("nil")
        }

        return enc.Encode(elem)

    default:
        return fmt.Errorf("invalid data type for encoding: %s",
            this_kind.String())
    }

    return nil
}

func (enc *Encoder) encode_map(v interface{}) (error) {
    m := reflect.ValueOf(v)
    keys := m.MapKeys()
    map_keys := make([]string, 0, len(keys))
    new_map := make(map[string]interface{}, len(keys))

    // keys in a map are required to be strings in bencode
    for _, k := range keys {
        skey := fmt.Sprintf("%s", k)
        map_keys = append(map_keys, skey)

        new_map[skey] = m.MapIndex(k).Interface()
    }

    // keys must be in lexical order
    sort.Strings(map_keys)

    w := enc.w
    w.Write([]byte{'d'})
    for _, k := range map_keys {
        err := enc.Encode(k)
        if err != nil {
            return err
        }

        err = enc.Encode(new_map[k])
        if err != nil {
            return err
        }

    }
    w.Write([]byte{'e'})


    return nil
}

func (enc *Encoder) encode_struct(v interface{}) (error) {

    
    return fmt.Errorf("encode_struct() not implemented yet: got '%+v'", v)
}

func (enc *Encoder) encode_slice(v interface{}) (error) {
    obj := reflect.ValueOf(v)

    w := enc.w
    w.Write([]byte{'l'})

    for i := 0; i < obj.Len(); i++ {
        err := enc.Encode(obj.Index(i).Interface())
        if err != nil {
            return err
        }
    }

    w.Write([]byte{'e'})

    return nil
}

func (enc *Encoder) encode_array(v interface{}) (error) {
    return enc.encode_slice(v)
}

// func Unmarshal(data []byte, v interface{}) error {
//     dec := NewDecoder(bytes.NewReader(data))
//     return dec.UnmarshalNext(v)
// }

// func (dec *Decoder) UnmarshalNext(v interface{}) error {

//     return nil
// }

func (dec *Decoder) Parse() (interface{}, error) {
    token, err := dec.Token()
    if err != nil {
        return nil, err
    }

    switch token.(type) {
    case Delim:
        switch token.(Delim) {
        case 'l':
            l, err := dec.parse_list()
            if err != nil {
                return nil, fmt.Errorf("error parsing list: %s", err)
            }
            return l, nil
        case 'd':
            d, err := dec.parse_dict()
            if err != nil {
                return nil, fmt.Errorf("error parsing dict: %s", err)
            }
            return d, nil
        default:
        }

    default:
        return token, nil
    }

    return nil, nil
}

func (dec *Decoder) parse_dict() (map[string]interface{}, error) {
    l, err := dec.parse_list()
    if err != nil {
        return nil, err
    }

    if (len(l) & 1) != 0 {
        return nil, fmt.Errorf("odd number of elements in dict at byte %d",
            dec.r.Tell())
    }

    d := make(map[string]interface{})
    for len(l) > 0 {
        k, ok := l[0].(string)
        if !ok {
            this_type := reflect.TypeOf(l[0])
            kind := this_type.Kind()
            return nil, fmt.Errorf("invalid type for dictionary key %s at byte %d.  must be a string.",
                kind.String(), dec.r.Tell())
        }
        d[k] = l[1]
        l = l[2:]
    }

    return d, nil
}

func (dec *Decoder) parse_list() ([]interface{}, error) {
    l := make([]interface{}, 0, 0)

    for token, err := dec.Token(); err == nil; token, err = dec.Token() {
        // switch t := token.(type) {
        switch token.(type) {
        case Delim:
            switch token.(Delim) {
            case 'l':
                this_list, err := dec.parse_list()
                if err != nil {
                    return nil, err
                }
                l = append(l, this_list)
            case 'd':
                this_dict, err := dec.parse_dict()
                if err != nil {
                    return nil, err
                }
                l = append(l, this_dict)
            case 'e':
                // end of list
                return l, nil
            default:
                return nil, fmt.Errorf("unrecognized token at byte %d",
                    dec.r.Tell())
            }

        default:
            l = append(l, token)
        }
    }

    return l, nil
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
        return nil, fmt.Errorf("unexpected byte '%s' near byte %d",
            s, r.Tell())
    }

    return nil, nil
}

func (dec *Decoder) get_string() (string, error) {
    size_64, err := dec.get_int(':')
    if err != nil {
        return "", err
    }
    size := int(size_64)
    if size < 0 {
        return "", fmt.Errorf("negative length specified for string at byte %s",
            dec.r.Tell())
    }

    p := make([]byte, size, size)
    p_read := p[:]
    amtread := 0

    r := dec.r
    for n, err := r.Read(p_read); amtread < size; n, err = r.Read(p_read) {
        amtread += n

        if err != nil {
            if err == io.EOF {
                break
            }
            return "", err
        }

        p_read = p_read[n:]
    }

    if amtread < size {
        return "", fmt.Errorf("short read while reading string")
    }

    return string(p), nil
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

        if (d >= '0' && d <= '9') || d == '-' {
            digits = append(digits, d)
            continue
        }

        if d == end {
            // done
            break
        }

        return 0, fmt.Errorf("unexpected byte '%s' in integer spec near byte %d",
            d, r.Tell())
    }

    return strconv.ParseInt(string(digits), 10, 64)
}

