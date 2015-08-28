package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fsouza/go-dockerclient"
	. "github.com/weaveworks/weave/common"
)

type createExecInterceptor struct{ proxy *Proxy }

func (i *createExecInterceptor) InterceptRequest(r *http.Request) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewReader(body))

	options := docker.CreateExecOptions{}
	if err := json.Unmarshal(body, &options); err != nil {
		return err
	}

	container, err := inspectContainerInPath(i.proxy.client, r.URL.Path)
	if err != nil {
		return err
	}

	_, hasWeaveWait := container.Volumes["/w"]
	cidrs, err := i.proxy.weaveCIDRsFromConfig(container.Config, container.HostConfig)
	if err != nil {
		Info.Printf("Ignoring container %s due to %s", container.ID, err)
	} else if hasWeaveWait {
		Info.Printf("Exec in container %s with WEAVE_CIDR \"%s\"", container.ID, strings.Join(cidrs, " "))
		cmd := append(weaveWaitEntrypoint, "-s")
		options.Cmd = append(cmd, options.Cmd...)

		if err := marshalRequestBody(r, options); err != nil {
			return err
		}
	}

	return nil
}

func (i *createExecInterceptor) InterceptResponse(r *http.Response) error {
	return nil
}
