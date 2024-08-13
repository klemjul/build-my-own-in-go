package internal

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type GitReference struct {
	Ref    string
	RefSha string
}

type RemoteRepository struct {
	BaseUrl    string
	httpClient http.Client
}

func NewRemoteRepository(BaseUrl string) (*RemoteRepository, error) {
	// TODO: remove test proxy
	proxyURL, err := url.Parse("http://localhost:8080")
	if err != nil {
		return nil, fmt.Errorf("error parsing proxy URL: %v", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := http.Client{
		Transport: transport,
	}
	return &RemoteRepository{
		httpClient: client,
		BaseUrl:    BaseUrl,
	}, nil
}

// https://git-scm.com/docs/gitprotocol-http/en#_discovering_references
func (r *RemoteRepository) DiscoveringReferences() ([]GitReference, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/info/refs?service=git-upload-pack", r.BaseUrl), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	res, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send req %v: %v", req.URL, err)
	}

	if strings.ToLower(res.Header.Get("Content-Type")) != "application/x-git-upload-pack-advertisement" {
		return nil, errors.New("clients SHOULD fall back to the dumb protocol if another content type is returned")
	}
	if res.StatusCode != 200 && res.StatusCode != 304 {
		return nil, errors.New("clients MUST validate the status code is either 200 OK or 304 Not Modified")
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body, %v", err)
	}

	lines := strings.Split(string(body), "\n")
	firstBytes := lines[0][:5]

	matched, err := regexp.MatchString("^[0-9a-f]{4}#", firstBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to regexp.MatchString, %v", err)
	}
	if !matched {
		return nil, errors.New("clients MUST validate the status code is either 200 OK or 304 Not Modified")
	}
	serviceName := strings.Split(strings.Split(lines[0], " ")[1], "=")[1]
	if serviceName != "git-upload-pack" {
		return nil, errors.New("clients MUST verify the first pkt-line is # service=$servicename")
	}

	refs := []GitReference{}
	for _, line := range lines[1:] {
		if strings.HasPrefix(line, "0000") {
			continue
		}
		ref := strings.Split(line, " ")
		if len(ref) > 1 {
			refs = append(refs, GitReference{Ref: ref[0], RefSha: ref[1]})
		}
	}
	return refs, nil
}
func Map[T any](slice []T, fn func(T) T) []T {
	result := make([]T, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// https://git-scm.com/docs/gitprotocol-http/en#_smart_service_git_upload_pack
//
// https://git-scm.com/docs/pack-protocol#_packfile_data
//
// https://git-scm.com/docs/pack-format
//
// https://bitbucket.org/ssaasen/git/src/master/Documentation/technical/pack-format.txt
//
// https://stefan.saasen.me/articles/git-clone-in-haskell-from-the-bottom-up/#implementing-ref-discovery
func (r *RemoteRepository) UploadPack(wants []string) error {
	reqBody := strings.Join(Map(wants, func(want string) string {
		return fmt.Sprintf("0032want %v", want[4:])
	}), "\n")
	reqBody += "\n00000009done\n"
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/git-upload-pack", r.BaseUrl), bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/x-git-upload-pack-request")
	req.Header.Set("Accept", "application/x-git-upload-pack-result")
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	res, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send req %v: %v", req.URL, err)
	}

	defer res.Body.Close()
	packType := make([]byte, 8)
	_, err = res.Body.Read(packType)
	if err != nil {
		return fmt.Errorf("failed to read body, %v", err)
	}

	packTypeExpected := []byte{'0', '0', '0', '8', 'N', 'A', 'K', '\n'}
	if !bytes.Equal(packType, packTypeExpected) {
		return fmt.Errorf("failed to parse pack, invalid header %v", string(packType))
	}
	packFileBytes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	headers, contents, _, err := parsePackFile(packFileBytes)
	if err != nil {
		panic(err)
	}
	// TODO: fix length issue
	for i := range 3 {
		fmt.Println(headers[i])
		fmt.Println(string(contents[i]))
	}

	return nil
}
