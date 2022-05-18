// This file is part of Go WML.
//
// Go WML is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Go WML is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Go WML.  If not, see <https://www.gnu.org/licenses/>.

package wml

import (
	"regexp"
	"strconv"
	"github.com/prataprc/goparsec"
	"strings"
	"fmt"
)

var float, _ = regexp.Compile(`^\d+\.\d+$`)
var integer, _ = regexp.Compile(`^\d+$`)

var openbracket = parsec.Atom ("[", "OPENBRACKET")
var tagname = parsec.TokenExact (`[a-zA-Z0-9_]+`, "TAGNAME")
var closebracket = parsec.Atom ("]", "CLOSEBRACKET")
var closetag = parsec.Atom ("[/", "CLOSETAG")
var textdom = parsec.Atom ("#textdomain", "TEXDOM")
var translated = parsec.Atom ("_", "TRANSLATED")

func one2one(ns []parsec.ParsecNode) parsec.ParsecNode {
	if ns == nil || len(ns) == 0 {
		return nil
	}
	return ns[0]
}

var textdomname = parsec.Token (`[a-zA-Z0-9_-]+`, "TEXTDOMNAME")
var textdomain = parsec.And (func (nodes[]parsec.ParsecNode) parsec.ParsecNode {
	name := nodes[1].(*parsec.Terminal).GetValue()
	return Domain{nil, name}//TOFIX	
}, textdom, textdomname)

func valueComponentNodify (ns[] parsec.ParsecNode) parsec.ParsecNode {
	if ns == nil || len(ns) != 1 {//do we really need this check?
		return nil
	}
	switch ns[0].(type) {
		case *parsec.Terminal://maybe raw string, maybe bool/int/float
			val := ns[0].(*parsec.Terminal).GetValue()
			/*switch {
				case val == "yes" || val == "true":
					return true
				case val == "no" || val == "false":
					return false
				case integer.MatchString(val):
					ival, _ := strconv.Atoi (val)
					return ival
				case float.MatchString(val):
					fval, _ := strconv.ParseFloat(val, 64)
					return fval
				default:*/
					return val
			//}
		case string://quoted string
			val := ns[0].(string)
			/*switch {
				case val == "yes" || val == "true":
					return true
				case val == "no" || val == "false":
					return false
				case integer.MatchString(val):
					ival, _ := strconv.Atoi (val)
					return ival
				case float.MatchString(val):
					fval, _ := strconv.ParseFloat(val, 64)
					return fval
				default:*/
					return val
			//}
		default://translated
			return ns[0]
	}
}

var wml_quoted_string_content = parsec.TokenExact(`(""|[^"])+`, "WMLQUOTEDSTRING")

/*
THIS WAS SUPER FLAWED!!! Strings were parsed one symbol by one, allocation A TERMINAL PER SYMBOL!!!
var wml_escaped_character = parsec.OrdChoice (one2one, parsec.AtomExact (`""`, "ESCAPED QUOTE"), parsec.TokenExact(`[^"]`, "COMMONSYMBOLS"))

var wml_quoted_string_content = parsec.Many (func (ns []parsec.ParsecNode) parsec.ParsecNode {
	var b strings.Builder
	for _, v := range ns {
		vTerm := v.(*parsec.Terminal)
		if vTerm.GetName() == "ESCAPED QUOTE" {
			b.WriteRune ('"')
		} else {
			b.WriteString (vTerm.GetValue())
		}
	}
	return b.String()
}, wml_escaped_character)*/

var wml_quoted_string = parsec.And (func (ns []parsec.ParsecNode) parsec.ParsecNode {
	val := ns[1].(*parsec.Terminal).GetValue()
	return strings.ReplaceAll (val, `""`, `"`)//we parsed with double quotes, now replace them as they should be
}, parsec.Atom (`"`, "QUOTE"),wml_quoted_string_content ,parsec.AtomExact (`"`, "QUOTEEXACT"))

var wml_translated_string = parsec.And (func (ns []parsec.ParsecNode) parsec.ParsecNode {
	stringNs := ns[2].(string)
	return Tr((stringNs))
}, parsec.Maybe (one2one, textdomain), translated, wml_quoted_string)//wml_quoted_string returns string, not Terminal

//TODO: raw strings with << >>
var wml_value_component = parsec.OrdChoice (valueComponentNodify, wml_translated_string, wml_quoted_string, parsec.Token (`[^+\n[]*`, "PLAIN"))

var wml_value = parsec.Many (func (ns []parsec.ParsecNode) parsec.ParsecNode {
	if len(ns) == 1 {
		return ns[0]
	}
	var b strings.Builder
	for _, v := range ns {//TODO: what if not string?
		switch val := v.(type) {
			case string: b.WriteString (val)
			//TODO: deletes Tr status, fix it
			case Tr: b.WriteString ((string)(val))
			case int: b.WriteString (strconv.Itoa (val))
			case bool: if val {
				b.WriteString ("yes")
			} else {
				b.WriteString ("no")
			}
			default: fmt.Println(v)
			return nil
		}
		
	}
	return b.String();	
	//return "TODO: Unsupported"//TODO: not supported for now, nasty descriptions
}, wml_value_component, parsec.Atom ("+", "PLUS"))

type tmpAttr struct {// temporary type for wml_attr parser
	key string
	val interface{}
}
var wml_attr = parsec.And (func (ns []parsec.ParsecNode) parsec.ParsecNode {
	return tmpAttr{key: ns[0].(*parsec.Terminal).GetValue(), val: ns[2]}
}, parsec.Token (`[a-zA-Z0-9_,]+`, "ATTR_KEY"), parsec.Atom ("=", "EQ"), wml_value)

var wml_data parsec.Parser
func tagNodify (ns []parsec.ParsecNode) parsec.ParsecNode {
	nameStart := ns[1].(*parsec.Terminal).GetValue()
	nameEnd := ns[5].(*parsec.Terminal).GetValue()
	if nameStart != nameEnd {//[name1][/name2], opening doesn't match closing
		return nil
	}
	return &Tag{Name: nameStart, Data: ns[3].(*Data)}
}
var wml_tag = parsec.And (tagNodify, openbracket, tagname, closebracket, &wml_data, closetag, tagname, closebracket)
var wml_datum = parsec.OrdChoice (one2one, textdomain, wml_attr, wml_tag)
func dataNodify (ns []parsec.ParsecNode) parsec.ParsecNode {
	data := NewData()
	lastDomain := -1
	for i, v := range ns {
		switch v.(type) {
			case Domain:
				lastDomain = i
				continue
			case tmpAttr:
				attr := v.(tmpAttr)
				switch attr.val.(type) {
					case Tr:
						if i != 0 && lastDomain == i - 1 {
							data.AddAttr (attr.key, Domain{V: attr.val, D: ns[lastDomain].(Domain).D})
						} else {
							data.AddAttr (attr.key, attr.val)
						}
					default:
						data.AddAttr (attr.key, attr.val)
				}
			default: //tag
				tag := v.(*Tag)
				data.AddTag (tag)
		}
	}
	return data
}
func init() {
	wml_data = parsec.Kleene (dataNodify, wml_datum)//had to use init(), it's a recursive definition
	//test, t2 := wml_data (parsec.NewScanner ([]byte("")))
	//fmt.Println ("Test", test, t2)
	//fmt.Printf ("%T\n", test)
	//dumb testing
	//test, t2 := wml_data (parsec.NewScanner ([]byte("[test]#textdomain fug\n[aza] ty = _ \"wer\"[/aza][/test]")))
	//fmt.Println ("Test", test, t2)
	//fmt.Printf ("%T\n", test)
	//test, t2 := wml_escaped_character (parsec.NewScanner ([]byte(`""`)))
	//fmt.Println ("Test", test, t2)
	//fmt.Printf ("%T\n", test)
}

func ParseTag(text []byte) *Tag {
	tag, _ := wml_tag (parsec.NewScanner (text))
	return tag.(*Tag)
}

func ParseData(text []byte) *Data {
	data, _ := wml_data (parsec.NewScanner (text))
	return data.(*Data)
}
