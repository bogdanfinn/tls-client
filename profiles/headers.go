package profiles

import (
	"fmt"
	"strconv"
	"strings"

	http "github.com/bogdanfinn/fhttp"
)

// The Chrome releases that changed what a navigation sends.
const (
	clientHintsSince = 89  // sec-ch-ua, sec-ch-ua-mobile and sec-ch-ua-platform on every request
	zstdSince        = 123 // zstd joined accept-encoding
	prioritySince    = 124 // the priority header, over HTTP/2 and HTTP/3
)

// DefaultHeaders returns the request headers the browser this profile imitates
// sends on a top-level navigation from Windows, in the order it sends them, as
// a set for WithDefaultHeaders. The second result is false for a profile whose
// browser is not mapped, and nothing is returned then, so a guess never goes on
// the wire.
//
// Chrome and Brave are mapped. The values and the order were read off the wire
// from Chrome, Edge and Brave 152 over HTTP/1.1 and HTTP/2, and the sec-ch-ua
// list is built the way Chromium builds it, see secChUA. What changed with a
// Chrome release follows the profile's major: zstd in accept-encoding from
// Chrome 123, the priority header from Chrome 124.
//
// The set is what the browser sends over HTTP/2, which this client negotiates
// by default. Over HTTP/1.1 the browser sends Host and Connection first, which
// the client writes for you, and no priority header; with WithForceHttp1,
// delete(headers, "priority") first. Names carry the browser's HTTP/1.1
// spelling, lower case for the client hints and capitals elsewhere, because
// HTTP/1.1 puts a name on the wire as it is written; HTTP/2 lower-cases every
// name.
//
// Firefox, Safari and Opera are not mapped: their header sets have not been
// read off the wire here, and a set written from memory is what this exists to
// replace.
func (c ClientProfile) DefaultHeaders() (http.Header, bool) {
	var brand string
	switch c.clientHelloId.Client {
	case "Chrome":
		brand = "Google Chrome"
	case "Brave":
		brand = "Brave"
	default:
		return nil, false
	}
	major, ok := majorOf(c.clientHelloId.Version)
	if !ok || major < clientHintsSince {
		return nil, false
	}
	brave := brand == "Brave"

	accept := "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"
	if !brave {
		// Brave does not offer signed exchanges.
		accept += ",application/signed-exchange;v=b3;q=0.7"
	}
	encoding := "gzip, deflate, br"
	if major >= zstdSince {
		encoding += ", zstd"
	}
	language := "en-US,en;q=0.9"
	if brave {
		// Brave changes the q value from one session to the next; 0.5, 0.7
		// and 0.8 were all seen from one build. Any of them is what a real
		// Brave sends.
		language = "en-US,en;q=0.7"
	}

	type field struct{ name, value string }
	fields := []field{
		{"Connection", "keep-alive"},
		{"sec-ch-ua", secChUA(major, brand)},
		{"sec-ch-ua-mobile", "?0"},
		{"sec-ch-ua-platform", `"Windows"`},
		{"Upgrade-Insecure-Requests", "1"},
		{"User-Agent", fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%d.0.0.0 Safari/537.36", major)},
		{"Accept", accept},
	}
	if brave {
		fields = append(fields, field{"Sec-GPC", "1"})
	}
	fields = append(fields,
		field{"Sec-Fetch-Site", "none"},
		field{"Sec-Fetch-Mode", "navigate"},
		field{"Sec-Fetch-User", "?1"},
		field{"Sec-Fetch-Dest", "document"},
		field{"Accept-Encoding", encoding},
		field{"Accept-Language", language},
	)
	if major >= prioritySince {
		fields = append(fields, field{"priority", "u=0, i"})
	}

	headers := make(http.Header, len(fields)+1)
	// Host is not in the set: the client writes it, over HTTP/1.1 through the
	// header map, so it takes the browser's place from the order.
	order := []string{"host"}
	for _, f := range fields {
		headers[f.name] = []string{f.value}
		order = append(order, strings.ToLower(f.name))
	}
	headers[http.HeaderOrderKey] = order
	return headers, true
}

// majorOf reads the leading number of a profile version, "152" as well as
// "152_PSK".
func majorOf(version string) (int, bool) {
	i := 0
	for i < len(version) && version[i] >= '0' && version[i] <= '9' {
		i++
	}
	n, err := strconv.Atoi(version[:i])
	return n, err == nil
}

// secChUA builds the sec-ch-ua value the way Chromium does, from
// components/embedder_support/user_agent_utils.cc. The list has three entries:
// an arbitrary brand, Chromium, and the browser's own brand. The arbitrary
// brand is "Not" + a character + "A" + a character + "Brand", with both
// characters and the version taken from tables indexed by the major, and the
// three entries are then placed by a permutation table indexed by the major
// too. Every part of the value therefore changes with the version, and a value
// copied from an older capture and bumped by hand keeps the wrong brand in the
// wrong place.
//
// Checked against what Brave 148 and 151 and Chrome, Edge and Brave 152 sent.
func secChUA(major int, brand string) string {
	chars := []string{" ", "(", ":", "-", ".", "/", ")", ";", "=", "?", "_"}
	versions := []string{"8", "99", "24"}
	orders := [][3]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}

	grease := "Not" + chars[major%len(chars)] + "A" + chars[(major+1)%len(chars)] + "Brand"
	entries := [3]string{
		fmt.Sprintf(`"%s";v="%s"`, grease, versions[major%len(versions)]),
		fmt.Sprintf(`"Chromium";v="%d"`, major),
		fmt.Sprintf(`"%s";v="%d"`, brand, major),
	}

	// The table says where each entry goes, not which entry comes next:
	// placed[order[i]] = entries[i]. Read the other way round it gives the
	// wrong order for four majors out of six, 148 among them.
	var placed [3]string
	for i, pos := range orders[major%len(orders)] {
		placed[pos] = entries[i]
	}
	return strings.Join(placed[:], ", ")
}
