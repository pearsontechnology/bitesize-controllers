package nginx

import (
    "fmt"
    "os"
    "os/exec"
    "io"
    "io/ioutil"

    "text/template"
)

var ConfigPath = "/etc/nginx"
var Command = "nginx"


func Start() error {
    nginxArgs := []string{
        "-c",
        ConfigPath + "/nginx.conf",
    }
    shellOut(Command, nginxArgs)
    return nil
}

func Verify() error {
    verifyArgs := []string{
        "-t",
        "-c",
        ConfigPath + "/nginx.conf",
    }
    return exec.Command(Command, verifyArgs...).Run()

}

func Reload() error {
    reloadArgs := []string{
        "-s",
        "reload",
    }
    return exec.Command(Command, reloadArgs...).Run()
}

func Template() (*template.Template, error) {
    return template.New("nginx.conf.tmpl").ParseFiles("nginx.conf.tmpl")

}

func WriteConfig(virtualHosts []*VirtualHost) error {
    // Needs to split into separate files
    fmt.Printf("Generating config\n")
    debug := os.Getenv("DEBUG")
    
    tmpl, err := Template()
    if err != nil {
        fmt.Printf("Error on template: %s\n", err.Error())
        return err
    }

    if w, err := os.Create(ConfigPath + "/nginx.conf"); err != nil {
        fmt.Printf("Error writing config: %s\n", err.Error())
        return err
    } else if err := tmpl.Execute(w, virtualHosts); err != nil {
        fmt.Printf("Error generating template: %s", err.Error())
        return err
    }

    if debug  == "true" {
        conf, _ := ioutil.ReadFile(ConfigPath + "/nginx.conf")
        fmt.Printf(string(conf))
    }

    return nil

}

func shellOut(shellCmd string, args []string) error {
  cmd := exec.Command(shellCmd, args...)
  stdout, _ := cmd.StdoutPipe()
  stderr, _ := cmd.StderrPipe()

  fmt.Printf("Starting %v %v\n", shellCmd, args)

  go io.Copy(os.Stdout, stdout)
  go io.Copy(os.Stderr, stderr)
  return cmd.Start()
}