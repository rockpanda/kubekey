package runner

import (
	"errors"
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"os"
)

type Runner struct {
	Conn  connector.Connection
	Debug bool
	Host  *kubekeyapiv1alpha1.HostCfg
	Index int
}

func (r *Runner) Exec(cmd string, printOutput bool) (string, string, int, error) {
	if r.Conn == nil {
		return "", "", 1, errors.New("no ssh connection available")
	}

	//stdout := NewTee(os.Stdout)
	//defer stdout.Close()
	//
	//stderr := NewTee(os.Stderr)
	//defer stderr.Close()
	//
	//code, err := r.Conn.PExec(cmd, nil, stdout, stderr)

	//if printOutput {
	//	if stdout.String() != "" {
	//		logger.Log.Infof("[stdout]: %s", stdout.String())
	//	}
	//	if stderr.String() != "" {
	//		logger.Log.Infof("[stderr]: %s", stderr.String())
	//	}
	//}
	//if err != nil {
	//	return "", err.Error(), code, err
	//}
	//
	//return stdout.String(), stderr.String(), code, nil
	stdout, stderr, code, err := r.Conn.Exec(cmd)
	if printOutput {
		if stdout != "" {
			logger.Log.Infof("[stdout]: %s", stdout)
		}
		if stderr != "" {
			logger.Log.Infof("[stderr]: %s", stderr)
		}
	}
	if err != nil {
		return "", stderr, code, err
	}
	return stdout, stderr, code, nil
}

func (r *Runner) Cmd(cmd string, printOutput bool) (string, error) {
	stdout, _, code, err := r.Exec(cmd, printOutput)
	if code != 0 || err != nil {
		return "", err
	}
	return stdout, nil
}

func (r *Runner) SudoExec(cmd string, printOutput bool) (string, string, int, error) {
	return r.Exec(ssh.SudoPrefix(cmd), printOutput)
}

func (r *Runner) SudoCmd(cmd string, printOutput bool) (string, error) {
	return r.Cmd(ssh.SudoPrefix(cmd), printOutput)
}

func (r *Runner) Fetch(local, remote string) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.Fetch(local, remote); err != nil {
		logger.Log.Errorf("fetch remote file %s to local %s failed: %v", remote, local, err)
		return err
	}
	logger.Log.Debugf("fetch remote file %s to local %s success", remote, local)
	return nil
}

func (r *Runner) Scp(local, remote string) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.Scp(local, remote); err != nil {
		logger.Log.Errorf("scp local file %s to remote %s failed: %v", local, remote, err)
		return err
	}
	logger.Log.Debugf("scp local file %s to remote %s success", local, remote)
	return nil
}

func (r *Runner) FileExist(remote string) (bool, error) {
	if r.Conn == nil {
		return false, errors.New("no ssh connection available")
	}

	ok := r.Conn.RemoteFileExist(remote)
	logger.Log.Debugf("check remote file exist: %v", ok)
	return ok, nil
}

func (r *Runner) DirExist(remote string) (bool, error) {
	if r.Conn == nil {
		return false, errors.New("no ssh connection available")
	}

	ok, err := r.Conn.RemoteDirExist(remote)
	if err != nil {
		logger.Log.Errorf("check remote dir exist failed: %v", err)
		return false, err
	}
	logger.Log.Debugf("check remote dir exist: %v", ok)
	return ok, nil
}

func (r *Runner) MkDir(path string) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.MkDirAll(path); err != nil {
		logger.Log.Errorf("make remote dir %s failed: %v", path, err)
		return err
	}
	return nil
}

func (r *Runner) Chmod(path string, mode os.FileMode) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.Chmod(path, mode); err != nil {
		logger.Log.Errorf("chmod remote path %s failed: %v", path, err)
		return err
	}
	return nil
}

func (r *Runner) FileMd5(path string) (string, error) {
	if r.Conn == nil {
		return "", errors.New("no ssh connection available")
	}

	cmd := fmt.Sprintf("md5sum %s | cut -d\" \" -f1", path)
	out, _, _, err := r.Conn.Exec(cmd)
	if err != nil {
		logger.Log.Errorf("count remote %s md5 failed: %v", path, err)
		return "", err
	}
	return out, nil
}

func SudoPrefix(cmd string) string {
	return fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", cmd)
}
