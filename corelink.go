package lwm2m

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrCoreLinkInvalidValue = errors.New("invalid core link string value")
)

// CoreLink is a link format use by Coap.
// Defines in RFC6690
// LWM2M use CoreLink to discover objects
type CoreLink struct {
	Uri    string
	Params map[string]string
}

func NewCoreLink() *CoreLink {
	return &CoreLink{
		Params: make(map[string]string),
	}
}

func (l *CoreLink) SetParam(key string, val string) {
	l.Params[key] = val
}

func (l *CoreLink) UnmarshalText(text []byte) error {
	if len(text) < 3 {
		return ErrCoreLinkInvalidValue
	}
	str := string(text)
	var elemRe = regexp.MustCompile(`<[^>]*>`)
	elemMatch := elemRe.FindString(str)
	if len(elemMatch) < 3 {
		return ErrCoreLinkInvalidValue
	}
	l.Uri = elemMatch[1 : len(elemMatch)-1]
	if len(text) > len(elemMatch) {
		attrs := strings.Split(str[len(elemMatch)+1:], ";")
		for _, attr := range attrs {
			pair := strings.Split(attr, "=")
			if len(pair) != 2 || len(pair[0]) == 0 {
				return ErrCoreLinkInvalidValue
			}
			l.Params[pair[0]] = strings.Replace(pair[1], "\"", "", -1)
		}
	}
	return nil
}

//	A CoRE resource discovery response may contains multiple CoreLink values
//
//    Link            = link-value-list
//    link-value-list = [ link-value *[ "," link-value ]]
//    link-value     = "<" URI-Reference ">" *( ";" link-param )
//    link-param     = ( ( "rel" "=" relation-types )
//                   / ( "anchor" "=" DQUOTE URI-Reference DQUOTE )
//                   / ( "rev" "=" relation-types )
//                   / ( "hreflang" "=" Language-Tag )
//                   / ( "media" "=" ( MediaDesc
//                          / ( DQUOTE MediaDesc DQUOTE ) ) )
//                   / ( "title" "=" quoted-string )
//                   / ( "title*" "=" ext-value )
//                   / ( "type" "=" ( media-type / quoted-mt ) )
//                   / ( "rt" "=" relation-types )
//                   / ( "if" "=" relation-types )
//                   / ( "sz" "=" cardinal )
//                   / ( link-extension ) )
//    link-extension = ( parmname [ "=" ( ptoken / quoted-string ) ] )
//                 / ( ext-name-star "=" ext-value )
//    ext-name-star  = parmname "*" ; reserved for RFC-2231-profiled
//                                  ; extensions.  Whitespace NOT
//                                  ; allowed in between.
//    ptoken         = 1*ptokenchar
//    ptokenchar     = "!" / "#" / "$" / "%" / "&" / "'" / "("
//                   / ")" / "*" / "+" / "-" / "." / "/" / DIGIT
//                   / ":" / "<" / "=" / ">" / "?" / "@" / ALPHA
//                   / "[" / "]" / "^" / "_" / "`" / "{" / "|"
//                   / "}" / "~"
//    media-type     = type-name "/" subtype-name
//    quoted-mt      = DQUOTE media-type DQUOTE
//    relation-types = relation-type
//                   / DQUOTE relation-type *( 1*SP relation-type ) DQUOTE
//    relation-type  = reg-rel-type / ext-rel-type
//    reg-rel-type   = LOALPHA *( LOALPHA / DIGIT / "." / "-" )
//    ext-rel-type   = URI
//    cardinal       = "0" / ( %x31-39 *DIGIT )
//    LOALPHA        = %x61-7A   ; a-z
//    quoted-string  = <defined in [RFC2616]>
//    URI            = <defined in [RFC3986]>
//    URI-Reference  = <defined in [RFC3986]>
//    type-name      = <defined in [RFC4288]>
//    subtype-name   = <defined in [RFC4288]>
//    MediaDesc      = <defined in [W3C.HTML.4.01]>
//    Language-Tag   = <defined in [RFC5646]>
//    ext-value      = <defined in [RFC5987]>
//    parmname       = <defined in [RFC5987]>

func CoreLinksFromString(s string) (links []*CoreLink, err error) {
	var re = regexp.MustCompile(`(<[^>]+>\s*(;\s*\w+\s*(=\s*(\w+|"([^"\\]*(\\.[^"\\]*)*)")\s*)?)*)`)
	m := re.FindAllString(s, -1)
	for _, match := range m {
		l := NewCoreLink()
		if err = l.UnmarshalText([]byte(match)); err == nil {
			links = append(links, l)
		}
	}
	return
}
