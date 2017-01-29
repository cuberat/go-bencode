

# bencode
`import "github.com/cuberat/go-bencode/bencode"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
The bencode package encodes/decodes data structures to/from the
Bencode format (<a href="https://en.wikipedia.org/wiki/Bencode">https://en.wikipedia.org/wiki/Bencode</a>).

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




## <a name="pkg-index">Index</a>
* [func Decode(r io.Reader) (interface{}, error)](#Decode)
* [func DecodeString(s string) (interface{}, error)](#DecodeString)
* [func Encode(w io.Writer, v interface{}) error](#Encode)
* [func EncodeToString(v interface{}) (string, error)](#EncodeToString)
* [type Decoder](#Decoder)
  * [func NewDecoder(r io.Reader) *Decoder](#NewDecoder)
  * [func (dec *Decoder) Decode() (interface{}, error)](#Decoder.Decode)
  * [func (dec *Decoder) Token() (Token, error)](#Decoder.Token)
* [type Delim](#Delim)
* [type Encoder](#Encoder)
  * [func NewEncoder(w io.Writer) *Encoder](#NewEncoder)
  * [func (enc *Encoder) Encode(v interface{}) error](#Encoder.Encode)
* [type Token](#Token)


#### <a name="pkg-files">Package files</a>
[bencode.go](/src/github.com/cuberat/go-bencode/bencode/bencode.go) 





## <a name="Decode">func</a> [Decode](/src/target/bencode.go?s=4784:4829#L166)
``` go
func Decode(r io.Reader) (interface{}, error)
```
Decode a Bencode data structure from the Reader, r.



## <a name="DecodeString">func</a> [DecodeString](/src/target/bencode.go?s=4201:4249#L141)
``` go
func DecodeString(s string) (interface{}, error)
```
Decode a Bencode data structure provided as a string, s.



## <a name="Encode">func</a> [Encode](/src/target/bencode.go?s=4626:4673#L160)
``` go
func Encode(w io.Writer, v interface{}) error
```
Encode the given data structure, v, to the Writer, w.



## <a name="EncodeToString">func</a> [EncodeToString](/src/target/bencode.go?s=4352:4402#L147)
``` go
func EncodeToString(v interface{}) (string, error)
```
Encode a data structure, v,  to a string.




## <a name="Decoder">type</a> [Decoder](/src/target/bencode.go?s=3397:3458#L105)
``` go
type Decoder struct {
    // contains filtered or unexported fields
}
```
Decoder object







### <a name="NewDecoder">func</a> [NewDecoder](/src/target/bencode.go?s=5612:5649#L213)
``` go
func NewDecoder(r io.Reader) *Decoder
```
Create a new Decoder to decode data structures from Bencode.





### <a name="Decoder.Decode">func</a> (\*Decoder) [Decode](/src/target/bencode.go?s=9493:9542#L378)
``` go
func (dec *Decoder) Decode() (interface{}, error)
```
Decode the Bencode data from the Reader provided to NewDecoder()
and return the resulting data structure as an interface.




### <a name="Decoder.Token">func</a> (\*Decoder) [Token](/src/target/bencode.go?s=12080:12122#L475)
``` go
func (dec *Decoder) Token() (Token, error)
```
Return the next Bencode token from the Reader provided to NewDecoder().
Return values are a Delim ('l', 'd', or 'e'), an int64, or a string.

You only need to worry about this if you want to handle decoding yourself.




## <a name="Delim">type</a> [Delim](/src/target/bencode.go?s=3691:3706#L119)
``` go
type Delim byte
```
A Delim is a byte representing the start or end of a list or dictionary:


	[ ] { }

You only need to worry about this if you want to handle decoding yourself.










## <a name="Encoder">type</a> [Encoder](/src/target/bencode.go?s=3478:3517#L111)
``` go
type Encoder struct {
    // contains filtered or unexported fields
}
```
Encoder object







### <a name="NewEncoder">func</a> [NewEncoder](/src/target/bencode.go?s=5451:5488#L205)
``` go
func NewEncoder(w io.Writer) *Encoder
```
Create a new Encoder to encode data structures to Bencode.





### <a name="Encoder.Encode">func</a> (\*Encoder) [Encode](/src/target/bencode.go?s=5814:5863#L222)
``` go
func (enc *Encoder) Encode(v interface{}) error
```
Encode the given data structure, v, to Bencode on the Writer provided to
NewEncoder().




## <a name="Token">type</a> [Token](/src/target/bencode.go?s=3995:4017#L129)
``` go
type Token interface{}
```
A Token holds a value of one of these types:


	Delim, representing the beginning lists and dictionaries: l d
	    or the end of one: e
	int64, for integers
	string, for strings

You only need to worry about this if you want to handle decoding yourself.














- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
