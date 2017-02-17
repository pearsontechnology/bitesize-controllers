package nginx

import (
    "os"
    "os/exec"
    "io"
    "io/ioutil"
    "strings"

    "text/template"
    log "github.com/Sirupsen/logrus"
)

var ConfigPath = "/etc/nginx"
var Command = "nginx"


func Start() error {
    nginxArgs := []string{
        "-c",
        ConfigPath + "/nginx.conf",
    }
    reloadArgs := []string{
        "-s",
        "reload",
    }
    searchArgs := strings.Join(nginxArgs," ")
    searchcmd := "ps -aux | grep \"" +  Command + " " + searchArgs + "\" | grep -v grep"
    err := exec.Command("sh","-c",searchcmd).Run()
    if err == nil {
        return exec.Command(Command, reloadArgs...).Run()
    } else {
        shellOut(Command, nginxArgs)
        return nil
    }

}

func Verify() error {
    verifyArgs := []string{
        "-t",
        "-c",
        ConfigPath + "/nginx.conf",
    }
    return exec.Command(Command, verifyArgs...).Run()

}

func Template() (*template.Template, error) {
    fm := template.FuncMap{
      "replace": func(str string, src string, dst string )  string {
        return strings.Replace(str, src, dst, -1)
      },
    }
    return template.New("nginx.conf.tmpl").Funcs(fm).ParseFiles("nginx.conf.tmpl")

}

func WriteConfig(virtualHosts []*VirtualHost) error {
    // Needs to split into separate files
    log.Info("Generating config")
    debug := os.Getenv("DEBUG")

    tmpl, err := Template()
    if err != nil {
        log.Errorf("Error on template: %s", err.Error())
        return err
    }

    if w, err := os.Create(ConfigPath + "/nginx.conf"); err != nil {
        log.Errorf("Error writing config: %s", err.Error())
        return err
    } else if err := tmpl.Execute(w, virtualHosts); err != nil {
        log.Errorf("Error generating template: %s", err.Error())
        return err
    }

    if debug  == "true" {
        conf, _ := ioutil.ReadFile(ConfigPath + "/nginx.conf")
        log.Debugf(string(conf))
    }

    return nil

}

func shellOut(shellCmd string, args []string) error {
  cmd := exec.Command(shellCmd, args...)
  stdout, _ := cmd.StdoutPipe()
  stderr, _ := cmd.StderrPipe()

  log.Infof("Starting %v %v\n", shellCmd, args)

  go io.Copy(os.Stdout, stdout)
  go io.Copy(os.Stderr, stderr)
  return cmd.Start()
}
