PACKAGE DOCUMENTATION

package bencode
    import "github.com/cuberat/go-bencode/bencode"

    The bencode package encodes/decodes data structures to/from the Bencode
    format (https://en.wikipedia.org/wiki/Bencode).

    Decoding:

	byte string -> string
	integer -> int64
	list -> []interface{}
	dictionary -> map[string]interface{}

    Encoding:

	string -> byte string
	int, int16, int32, int64 -> integer
	float32, float64 -> byte string
	any slice -> list
	map -> dictionary
	struct -> dictionary

    Examples:

	package main

	import (
	    "fmt"
	    "github.com/cuberat/go-bencode/bencode"
	    "log"
	)

	func main() {
	    bencode_str := "d3:bar4:spam4:catsli1ei-2ei3ee3:fooi42ee"

	    data, err := bencode.DecodeString(bencode_str)
	    if err != nil {
	        log.Fatalf("couldn't decode: %s", err)
	    }

	    fmt.Printf("decoded %+v\n", data)
	}
	----
	>$ decoded map[cats:[1 -2 3] foo:42 bar:spam]

	--------------------

	package main

	import (
	    "fmt"
	    "github.com/cuberat/go-bencode/bencode"
	    "log"
	)

	func main() {
	    my_map := map[string]string{}
	    my_map["foo"] = "bar2"
	    my_map["bar"] = "foo2"
	    my_map["zebra"] = "not a horse"

	    encoded, err := bencode.EncodeToString(my_map)
	    if err != nil {
	        log.Fatalf("couldn't encode: %s", err)
	    }

	    fmt.Printf("encoded data as '%s'\n", encoded)
	}
	----
	>$ encoded data as 'd3:bar4:foo23:foo4:bar25:zebra11:not a horsee'

FUNCTIONS

func Decode(r io.Reader) (interface{}, error)
    Decode a Bencode data structure from the Reader, r.

func DecodeString(s string) (interface{}, error)
    Decode a Bencode data structure provided as a string, s.

func Encode(w io.Writer, v interface{}) error
    Encode the given data structure, v, to the Writer, w.

func EncodeToString(v interface{}) (string, error)
    Encode a data structure, v, to a string.

TYPES

type Decoder struct {
    // contains filtered or unexported fields
}
    Decoder object

func NewDecoder(r io.Reader) *Decoder
    Create a new Decoder to decode data structures from Bencode.

func (dec *Decoder) Decode() (interface{}, error)
    Decode the Bencode data from the Reader provided to NewDecoder() and
    return the resulting data structure as an interface.

func (dec *Decoder) Token() (Token, error)
    Return the next Bencode token from the Reader provided to NewDecoder().
    Return values are a Delim ('l', 'd', or 'e'), an int64, or a string.

    You only need to worry about this if you want to handle decoding
    yourself.

type Delim byte
    A Delim is a byte representing the start or end of a list or dictionary:

	[ ] { }

    You only need to worry about this if you want to handle decoding
    yourself.

type Encoder struct {
    // contains filtered or unexported fields
}
    Encoder object

func NewEncoder(w io.Writer) *Encoder
    Create a new Encoder to encode data structures to Bencode.

func (enc *Encoder) Encode(v interface{}) error
    Encode the given data structure, v, to Bencode on the Writer provided to
    NewEncoder().

type Token interface{}
    A Token holds a value of one of these types:

	Delim, representing the beginning lists and dictionaries: l d
	    or the end of one: e
	int64, for integers
	string, for strings

    You only need to worry about this if you want to handle decoding
    yourself.


