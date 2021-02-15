package device

import (
	"regexp"
	"strings"
)

type CoreLink struct {
	URI string
	Params map[string]interface{}
}

func NewCoreLink() *CoreLink {
	return &CoreLink{
		Params: make(map[string]interface{}),
	}
}

func (l *CoreLink) SetParam(key string, val interface{})  {
	l.Params[key] = val
}

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

func CoreLinksFromString(s string) []*CoreLink {

	var re = regexp.MustCompile(`(<[^>]+>\s*(;\s*\w+\s*(=\s*(\w+|"([^"\\]*(\\.[^"\\]*)*)")\s*)?)*)`)
	var elemRe = regexp.MustCompile(`<[^>]*>`)

	var links []*CoreLink
	m := re.FindAllString(s, -1)

	for _, match := range m {
		elemMatch := elemRe.FindString(match)
		l := NewCoreLink()
		l.URI = elemMatch[1 : len(elemMatch)-1]
		if len(match) > len(elemMatch) {
			attrs := strings.Split(match[len(elemMatch)+1:], ";")
			for _, attr := range attrs {
				pair := strings.Split(attr, "=")
				l.Params[pair[0]] = strings.Replace(pair[1], "\"", "", -1)
			}
		}
		links = append(links, l)
	}

	return links
}