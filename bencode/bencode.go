// BSD 2-Clause License
//
// Copyright (c) 2017 Don Owens <don@regexguy.com>.  All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.


// The bencode package encodes/decodes data structures to/from the
// Bencode format (https://en.wikipedia.org/wiki/Bencode).
//
// Decoding:
//   byte string -> string
//   integer -> int64
//   list -> []interface{}
//   dictionary -> map[string]interface{}
//
// Encoding:
//   string -> byte string
//   int, int16, int32, int64 -> integer
//   float32, float64 -> byte string
//   any slice -> list
//   map -> dictionary
//   struct -> dictionary
//
// Examples:
//
//    package main
//
//    import (
//        "fmt"
//        "github.com/cuberat/go-bencode/bencode"
//        "log"
//    )
//
//    func main() {
//        bencode_str := "d3:bar4:spam4:catsli1ei-2ei3ee3:fooi42ee"
//
//        data, err := bencode.DecodeString(bencode_str)
//        if err != nil {
//            log.Fatalf("couldn't decode: %s", err)
//        }
//
//        fmt.Printf("decoded %+v\n", data)
//    }
//    ----
//    >$ decoded map[cats:[1 -2 3] foo:42 bar:spam]
//
//    --------------------
//
//    package main
//
//    import (
//        "fmt"
//        "github.com/cuberat/go-bencode/bencode"
//        "log"
//    )
//
//    func main() {
//        my_map := map[string]string{}
//        my_map["foo"] = "bar2"
//        my_map["bar"] = "foo2"
//        my_map["zebra"] = "not a horse"
//
//        encoded, err := bencode.EncodeToString(my_map)
//        if err != nil {
//            log.Fatalf("couldn't encode: %s", err)
//        }
//
//        fmt.Printf("encoded data as '%s'\n", encoded)
//    }
//    ----
//    >$ encoded data as 'd3:bar4:foo23:foo4:bar25:zebra11:not a horsee'
package bencode

import (
    "bufio"
    "bytes"
    "fmt"
    // "log"
    "io"
    "os"
    "reflect"
    "sort"
    "strconv"
    "strings"
)

const Version = "0.9.2"

// Decoder object
type Decoder struct {
    // r *bufio.Reader
    r *breader
}

// Encoder object
type Encoder struct {
    w io.Writer
}

// A Delim is a byte representing the start or end of a list or dictionary:
//     [ ] { }
//
// You only need to worry about this if you want to handle decoding yourself.
type Delim byte

// A Token holds a value of one of these types:
//
//     Delim, representing the beginning lists and dictionaries: l d
//         or the end of one: e
//     int64, for integers
//     string, for strings
//
// You only need to worry about this if you want to handle decoding yourself.
type Token interface{}

// func Unmarshal(data []byte, v interface {}) error {

// }

type breader struct {
    r *bufio.Reader
    pos uint64
}

func FillData(out_intfc interface{}, in_intfc interface{}) error {
    out := reflect.ValueOf(out_intfc)
    in := reflect.ValueOf(in_intfc)

    k := out.Kind()
    if k == reflect.Interface {
        out = out.Elem()

        if ! out.IsValid() {
            return fmt.Errorf("invalid value passed to decoder")
        }
        k = out.Kind()
    }
    if k == reflect.Ptr {
        out = out.Elem()
        if ! out.IsValid() {
            return fmt.Errorf("invalid value passed to decoder")
        }
        k = out.Kind()
    }

    return set_val_coerce(&out, in)
}

func unmarshal_struct(out *reflect.Value, in reflect.Value) (error) {
    d, ok := in.Interface().(map[string]interface{})
    if !ok {
        return fmt.Errorf("FillData not passed map[string]interface{}")
    }

    t := out.Type()

    for i := 0; i < t.NumField(); i++ {
        f := t.Field(i)
        tag_val := f.Tag.Get("bencode")
        flag_list := strings.Split(tag_val, ",")
        name := flag_list[0]
        if name == "" {
            name = f.Name
        }

        d_data, ok := d[name]
        if ok {
            f_val := out.Field(i)
            d_val := reflect.ValueOf(d_data)
            // fk := f_val.Kind()
            // d_k := d_val.Kind()
            // fmt.Fprintf(os.Stderr, "setting field %s (%s), input is a %s\n", name, fk, d_k)
            // f_val.Set(reflect.ValueOf(d_data))

            err := set_val_coerce(&f_val, d_val)
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func set_val_coerce(out *reflect.Value, in reflect.Value) error {
    out_kind := out.Kind()
    out_type := out.Type()
    in_kind := in.Kind()
    in_type := in.Type()

    if in_kind == out_kind {
        if in_type == out_type {
            out.Set(in)
            return nil
        }
    }

    if out_kind == reflect.Interface {
        out.Set(reflect.ValueOf(in.Interface()))
        return nil
    } else {
        if in_kind == reflect.Interface {
            new_in := in.Elem()
            return set_val_coerce(out, new_in)
        }
    }


    switch {
    case out_kind == reflect.String:
        return set_val_coerce_to_string(out, in)
    case is_kind_int(out_kind):
        return set_val_coerce_to_int(out, in)
    case is_kind_float(out_kind):
        return set_val_coerce_to_float(out, in)
    case out_kind == reflect.Struct:
        return unmarshal_struct(out, in)
    case out_kind == reflect.Slice:
        return set_val_coerce_slice(out, in)

    }

    return fmt.Errorf("don't know how to coerce %s to %s (%s to %s) (%T to %T)",
        in.Kind(), out.Kind(), in.Type(), out.Type(), in, out)
}

func set_val_coerce_slice(out *reflect.Value, in reflect.Value) error {
    in_type := in.Type()
    out_type := out.Type()
    in_kind := in.Kind()
    // out_kind := out.Kind()

    in_length := in.Len()
    // cap := in_length

    if in_kind != reflect.Slice {
        if in_kind == reflect.String {
            if _, ok := in.Interface().([]byte); ok {
                s := in.String()
                out.Set(reflect.ValueOf([]byte(s)))
                return nil
            }
        }
        // FIXME: stringify?

        return fmt.Errorf("don't know how to coerce %T to %T",
            in.Interface(), out.Interface())
    }

    out_elem_type := out_type.Elem()

    if in_length == 0 {
        new_in := reflect.MakeSlice(out_elem_type, 0, 0)
        out.Set(new_in)
        return nil
    }

    if in_type == out_type {
        fmt.Fprintf(os.Stderr, "in_type == out_type: %s = %s\n", in_type, out_type)
        out.Set(in)
        // reflect.SliceOf(type)
        // slice_type :=

        return nil
    }


    new_in := reflect.MakeSlice(out_type, 0, in_length)

    for i := 0; i < in_length; i++ {
        elem := in.Index(i)
        new_val_ptr := reflect.New(out_elem_type)
        new_val := new_val_ptr.Elem()

        err := set_val_coerce(&new_val, elem)
        if err != nil {
            return fmt.Errorf("couldn't coerce %T(%s) to %T(%s) in slice",
                elem.Interface(), elem.Kind(), new_val.Interface(), new_val.Kind())
        }

        new_in = reflect.Append(new_in, new_val)
    }

    out.Set(new_in)

    return nil

    // return fmt.Errorf("don't know how to coerce %T to %T",
    //     in.Interface(), out.Interface())
}

func set_val_coerce_to_string(out *reflect.Value, in reflect.Value) error {
    in_kind := in.Kind()

    if in_kind == reflect.String {
        out.Set(in)
        return nil
    }

    if is_signed, ok := get_int_kind(in_kind); ok {
        var s string
        if is_signed {
            s = strconv.FormatInt(in.Int(), 10)
        } else {
            strconv.FormatUint(in.Uint(), 10)
        }
        out.SetString(s)

        return nil
    }

    if is_kind_float(in_kind) {
        s := strconv.FormatFloat(in.Float(), 'g', -1, 64)
        out.SetString(s)
        return nil
    }

    if in_kind == reflect.Slice {
        intfc := in.Interface()
        if byte_slice, ok := intfc.([]byte); ok {
            s := string(byte_slice)
            out.SetString(s)
            return nil
        }
    }

    return fmt.Errorf("don't know how to coerce %s to %s (%s to %s) (%T to %T)",
        in.Kind(), out.Kind(), in.Type(), out.Type(),
        in.Interface(), out.Interface())
}

func set_val_coerce_to_float(out *reflect.Value, in reflect.Value) error {
    in_kind := in.Kind()
    if is_kind_float(in_kind) {
        out.SetFloat(in.Float())
        return nil
    }

    if in_kind == reflect.String {
        in_float, err := strconv.ParseFloat(in.String(), 64)
        if err != nil {
            return err
        }
        out.SetFloat(in_float)
        return nil
    }

    if is_signed, ok := get_int_kind(in_kind); ok {
        var in_float float64
        if is_signed {
            in_float = float64(in.Int())
        } else {
            in_float = float64(in.Uint())
        }
        out.SetFloat(in_float)
        return nil
    }

    return fmt.Errorf("don't know how to coerce %s to %s (%s to %s)",
        in.Kind(), out.Kind(), in.Type(), out.Type())
}

func set_val_coerce_to_int(out *reflect.Value, in reflect.Value) error {
    out_kind := out.Kind()
    out_type := out.Type()
    in_kind := in.Kind()
    in_type := in.Type()

    if in_kind == out_kind {
        out.Set(in)
        return nil
    }

    if in_is_signed, ok := get_int_kind(in_kind); ok {
        return set_val_coerce_int_to_int(out, in, in_is_signed)
    }

    switch in_kind {
    case reflect.String:
        return set_val_coerce_string_to_int(out, in)
    }

    return fmt.Errorf("don't know how to coerce %s to %s (%s to %s)",
        in_kind, out_kind, in_type, out_type)
}

func set_val_coerce_int_to_int(out *reflect.Value, in reflect.Value,
    in_is_signed bool) error {

    switch out.Kind() {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        if in_is_signed {
            out.SetInt(in.Int())
        } else {
            out.SetInt(int64(in.Uint()))
        }

    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        if in_is_signed {
            out.SetUint(uint64(in.Int()))
        } else {
            out.SetUint(in.Uint())
        }

    default:
        return fmt.Errorf("don't know how to coerce %s to %s (%s to %s)",
            in.Kind(), out.Kind(), in.Type(), out.Type())
    }

    return nil
}

func set_val_coerce_string_to_int(out *reflect.Value, in reflect.Value) error {
    switch out.Kind() {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        the_int, err := strconv.ParseInt(in.String(), 10, 64)
        if err != nil {
            return err
        }
        out.SetInt(the_int)

    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        the_uint, err := strconv.ParseUint(in.String(), 10, 64)
        if err != nil {
            return err
        }
        out.SetUint(the_uint)

    default:
        return fmt.Errorf("don't know how to coerce %s to %s (%s to %s)",
            in.Kind(), out.Kind(), in.Type(), out.Type())
    }

    return nil
}

func is_kind_float(kind reflect.Kind) bool {
    switch kind {
    case reflect.Float32, reflect.Float64:
        return true
    }

    return false
}

func is_kind_int(kind reflect.Kind) bool {
    switch kind {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        fallthrough
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        return true
    }

    return false
}

func get_int_kind(kind reflect.Kind) (is_signed, ok bool) {
    is_signed = false
    ok = false
    switch kind {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        is_signed = true
        ok = true
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        ok = true
    }

    return is_signed, ok
}

// Decode a Bencode data structure provided as a string, s.
func DecodeString(s string) (interface{}, error) {
    r := strings.NewReader(s)
     return Decode(r)
}

// Encode a data structure, v,  to a string.
func EncodeToString(v interface{}) (string, error) {
    buf := new(bytes.Buffer)
    enc := NewEncoder(buf)

    err := enc.Encode(v)
    if err != nil {
        return "", err
    }

    return buf.String(), nil
}

// Encode the given data structure, v, to the Writer, w.
func Encode(w io.Writer, v interface{}) (error) {
    enc := NewEncoder(w)
    return enc.Encode(v)
}

// Decode a Bencode data structure from the Reader, r.
func Decode(r io.Reader) (interface{}, error) {
    dec := NewDecoder(r)
    v, err := dec.Decode()

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

// Create a new Encoder to encode data structures to Bencode.
func NewEncoder(w io.Writer) *Encoder {
    enc := new(Encoder)
    enc.w = w

    return enc
}

// Create a new Decoder to decode data structures from Bencode.
func NewDecoder(r io.Reader) *Decoder {
    dec := new(Decoder)
    dec.r = new_reader(r)

    return dec
}

// Encode the given data structure, v, to Bencode on the Writer provided to
// NewEncoder().
func (enc *Encoder) Encode(v interface{}) (error) {
    vt, ok := v.(reflect.Value)
    if ok {
        v = vt.Interface()
    }

    this_type := reflect.TypeOf(v)
    this_kind := this_type.Kind()

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
    t := reflect.TypeOf(v)
    val := reflect.ValueOf(v)

    field_map := make(map[string]interface{}, val.NumField())

    for i := 0; i < t.NumField(); i++ {
        f := t.Field(i)
        fv := val.Field(i)

        field_map[f.Name] = fv
    }

    return enc.encode_map(field_map)
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


// Decode the Bencode data from the Reader provided to NewDecoder()
// and return the resulting data structure as an interface.
func (dec *Decoder) Decode() (interface{}, error) {
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

// Return the next Bencode token from the Reader provided to NewDecoder().
// Return values are a Delim ('l', 'd', or 'e'), an int64, or a string.
//
// You only need to worry about this if you want to handle decoding yourself.
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
