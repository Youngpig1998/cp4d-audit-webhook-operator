// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// IBM Confidential
// Automated Tests
// Copyright IBM Corp. 2021
// ------------------------------------------------------ {COPYRIGHT-END} ---
package testing

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// DoesFileExist returns if a file exists
func DoesFileExist(target string) bool {
	if _, err := os.Stat(target); err == nil {
		return true
	}
	return false
}

// IsFileEmpty returns if a file has any contents
func IsFileEmpty(target string) (bool, error) {
	if !DoesFileExist(target) {
		return true, errors.Errorf("file '%s' does not exist, so cannot check contents", target)
	}
	content, err := ioutil.ReadFile(target)
	if err != nil {
		return true, err
	}

	if len(content) > 0 {
		return false, nil
	}
	return true, nil
}

// IsStringInFile returns if a desired file contains a desired string
func IsStringInFile(targetFile, targetString string) (bool, error) {
	if !DoesFileExist(targetFile) {
		return false, errors.Errorf("file '%s' does not exist, so cannot check contents", targetFile)
	}
	contents, err := ioutil.ReadFile(targetFile)
	if err != nil {
		return false, err
	}
	search := string(contents)
	return strings.Contains(search, targetString), nil
}

// DeleteFile deletes a desired file if it exists
func DeleteFile(target string) error {
	if !DoesFileExist(target) {
		return nil
	}
	return os.Remove(target)
}

// CopyPod defines a particular container in a pod
type CopyPod struct {
	Name          string
	Namespace     string
	ContainerName string
}

// InitRestClient returns a ClientConfig
func InitRestClient() (*rest.Config, *corev1client.CoreV1Client, error) {
	// Instantiate loader for kubeconfig file
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	// Get a rest.Config from the kubeconfig file.  This will be passed into all
	// the client objects we create.
	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, nil, err
	}
	// Create a Kubernetes core/v1 client.
	coreclient, err := corev1client.NewForConfig(restconfig)
	if err != nil {
		return nil, nil, err
	}
	return restconfig, coreclient, err
}

// CopyFromPod copies out a desired file from the pod, and writes it locally onto the machine
// in a desired destination
func (i *CopyPod) CopyFromPod(srcPath string, destPath string) error {
	restconfig, coreclient, err := InitRestClient()
	if err != nil {
		return err
	}
	reader, outStream := io.Pipe()
	cmdArr := []string{"tar", "cf", "-", srcPath}
	req := coreclient.RESTClient().
		Get().
		Namespace(i.Namespace).
		Resource("pods").
		Name(i.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: i.ContainerName,
			Command:   cmdArr,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restconfig, "POST", req.URL())
	if err != nil {
		return err
	}

	go func() {
		defer outStream.Close()
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: outStream,
			Stderr: os.Stderr,
			Tty:    false,
		})
		// can't get the err out of this go func
		// tried channels but it still hangs
		// but we check futher down if the file has
		// actually been created before returning out happily
		if err != nil {
			fmt.Println("error occured when copying file from pod")
		}
	}()

	prefix := getPrefix(srcPath)
	prefix = path.Clean(prefix)
	prefix = stripPathShortcuts(prefix)
	destPath = path.Join(destPath, path.Base(prefix))
	err = untarAll(reader, destPath, prefix)
	if err != nil {
		return err
	}
	// Will no longer need to do this if we're checking in memory
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		return fmt.Errorf("there was a failure in the copy: %s does not exist when it's supposed to by now", destPath)
	}
	return err
}

// untarAll untars the file(s) being extracted from the pod and writes the file(s) to a desired directory
func untarAll(reader io.Reader, destDir, prefix string) error {
	tarReader := tar.NewReader(reader)
	for {
		// Next advances to the next entry in the tar archive
		header, err := tarReader.Next()
		if err != nil {
			// io.EOF is returned at the end of the input
			if err != io.EOF {
				return err
			}
			break
		}

		if !strings.HasPrefix(header.Name, prefix) {
			return fmt.Errorf("tar contents corrupted")
		}

		mode := header.FileInfo().Mode()
		destFileName := filepath.Join(destDir, header.Name[len(prefix):])

		baseName := filepath.Dir(destFileName)
		if err := os.MkdirAll(baseName, 0755); err != nil {
			return err
		}
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(destFileName, 0755); err != nil {
				return err
			}
			continue
		}

		evaledPath, err := filepath.EvalSymlinks(baseName)
		if err != nil {
			return err
		}

		if mode&os.ModeSymlink != 0 {
			linkname := header.Linkname

			if !filepath.IsAbs(linkname) {
				_ = filepath.Join(evaledPath, linkname)
			}

			if err := os.Symlink(linkname, destFileName); err != nil {
				return err
			}
			// Eventually modify this to read in memory rather than write to FS
			// Could also include a feature for reading through to find a target string
		} else {
			outFile, err := os.Create(destFileName)
			if err != nil {
				return err
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

func getPrefix(file string) string {
	return strings.TrimLeft(file, "/")
}

// stripPathShortcuts removes any leading or trailing "../" from a given path
func stripPathShortcuts(p string) string {
	newPath := path.Clean(p)
	trimmed := strings.TrimPrefix(newPath, "../")

	for trimmed != newPath {
		newPath = trimmed
		trimmed = strings.TrimPrefix(newPath, "../")
	}

	// trim leftover {".", ".."}
	if newPath == "." || newPath == ".." {
		newPath = ""
	}

	if len(newPath) > 0 && string(newPath[0]) == "/" {
		return newPath[1:]
	}
	return newPath
}
