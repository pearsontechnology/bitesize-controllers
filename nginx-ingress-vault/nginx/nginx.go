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
    fm := template.FuncMap{
      "replace": func(str string, src string, dst string )  string {
        return strings.Replace(str, src, dst, -1)
      },
    }
    return template.New("nginx.conf.tmpl").Funcs(fm).ParseFiles("nginx.conf.tmpl")

}

func WriteCustomErrorPages(virtualHosts []*VirtualHost) error {
    // cops-165 - Generate custom error page per vhost
    etmpl, _ := template.ParseFiles("error_page.tmpl")
    log.Info("load the template file for custom_error_pages")

    for host := range virtualHosts {
        if epage, err := os.Create("/usr/share/nginx/html/error_" + virtualHosts[host].Name + ".html"); err != nil {
            log.Errorf("Error writing config: %s", err.Error())
            return err
        } else if err := etmpl.Execute(epage, virtualHosts[host]); err != nil {
            log.Errorf("Error creating custom_error_page: error_" + virtualHosts[host].Name + ".html")
            return err
        } else {
            log.Info("Create custom_error_page: error_" + virtualHosts[host].Name + ".html")
            epage.Close()
        }
    }

    return nil

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
