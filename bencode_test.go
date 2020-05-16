package bencode_test

import (
    bencode "github.com/cuberat/go-bencode"
    "fmt"
    "reflect"
    "testing"
)

type TestItem struct {
    Encoded string
    Decoded interface{}
}

func TestDecode(t *testing.T) {
    test_data := get_test_data()

    for idx, item := range test_data {
        name := fmt.Sprintf("%d - %s", idx, item.Encoded)
        t.Run(name, func(st *testing.T) {
            got, err := bencode.DecodeString(item.Encoded)
            if err != nil {
                t.Errorf("error decoding string: %s", err)
                return
            }

            if !reflect.DeepEqual(got, item.Decoded) {
                t.Errorf("got %v, expected %v", got, item.Decoded)
                return
            }
        })
    }
}

func TestEncode (t *testing.T) {
    test_data := get_test_data()

    for idx, item := range test_data {
        name := fmt.Sprintf("%d - %s", idx, item.Encoded)
        t.Run(name, func(st *testing.T) {
            got, err := bencode.EncodeToString(item.Decoded)
            if err != nil {
                t.Errorf("error encoding data: %s", err)
                return
            }

            if got != item.Encoded {
                t.Errorf("got %q, expected %q", got, item.Encoded)
                return
            }
        })
    }
}

func get_test_data() ([]*TestItem) {
    return []*TestItem{
        // Examples from https://en.wikipedia.org/wiki/Bencode
        &TestItem{"i42e", int64(42)},
        &TestItem{"i0e", int64(0)},
        &TestItem{"i-42e", int64(-42)},
        &TestItem{"l4:spami42ee", []interface{}{"spam", int64(42)}},
        &TestItem{"d3:bar4:spam3:fooi42ee",
            map[string]interface{}{"bar": "spam", "foo": int64(42)},
        },

        // Examples taken from spec from theory.org:
        // https://wiki.theory.org/index.php/BitTorrentSpecification#Byte_Strings
        &TestItem{"4:spam", "spam"},
        &TestItem{"0:", ""},
        &TestItem{"i3e", int64(3)},
        &TestItem{"i-3e", int64(-3)},
        &TestItem{"l4:spam4:eggse", []interface{}{"spam","eggs"}},
        &TestItem{"le", []interface{}{}},
        &TestItem{"d3:cow3:moo4:spam4:eggse",
            map[string]interface{}{"cow": "moo", "spam": "eggs"},
        },
        &TestItem{"d4:spaml1:a1:bee",
            map[string]interface{}{"spam": []interface{}{"a", "b"}},
        },
        &TestItem{"d9:publisher3:bob17:publisher-webpage15:" +
            "www.example.com18:publisher.location4:homee",
            map[string]interface{}{"publisher": "bob",
                "publisher-webpage": "www.example.com",
                "publisher.location": "home"},
        },
        &TestItem{"de", map[string]interface{}{}},

        // From the Perl Bencode module:
        // https://metacpan.org/source/ARISTOTLE/Bencode-1.501/t/01-bdecode.t
        &TestItem{"d8:spam.mp3d6:author5:Alice6:lengthi100000eee",
            map[string]interface{}{
                "spam.mp3": map[string]interface{}{
                    "author":"alice", "length": "100000",
                },
            },
        },
    }
}
