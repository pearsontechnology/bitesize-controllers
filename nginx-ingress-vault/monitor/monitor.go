package monitor

import (
    "github.com/prometheus/client_golang/prometheus"
)

type MonitorCounter struct {
    desc     *prometheus.Desc
    counter  float64
}
type Monitor struct {
    Version           string
    VHosts            MonitorCounter
    SslVHosts         MonitorCounter
    NonSslVHosts      MonitorCounter
    FailedVHosts      MonitorCounter
    FailedSslVHosts   MonitorCounter
    NoCertSslVHosts   MonitorCounter
    SslVHostsCertFail MonitorCounter
    ConfigErrors      MonitorCounter
    TemplateErrors    MonitorCounter
}
const version = "1.9.11"

var Status Monitor

func init() {
    Status.Version = version

    Status.VHosts.desc = prometheus.NewDesc("ingress_vhosts",
        "Current number of ingress virtual hosts defined", nil, nil,)

    Status.SslVHosts.desc = prometheus.NewDesc("ingress_vhosts_ssl",
        "Current number ingresses with ssl true", nil, nil,)

    Status.NonSslVHosts.desc = prometheus.NewDesc("ingress_vhosts_non_ssl",
        "Current number ingresses withuot ssl true", nil, nil,)

    Status.FailedVHosts.desc = prometheus.NewDesc("ingress_vhosts_failed",
        "Current number of ingresses that failed to be rendered as virtual hosts", nil, nil,)

    Status.FailedSslVHosts.desc = prometheus.NewDesc("ingress_vhosts_failed_ssl",
        "Current number of ingresses that failed to be rendered as ssl virtual hosts", nil, nil,)

    Status.NoCertSslVHosts.desc = prometheus.NewDesc("ingress_vhosts_ssl_no_cert",
        "Current number of ingresses with ssl true but no cert found in vault", nil, nil,)

    Status.SslVHostsCertFail.desc = prometheus.NewDesc("ingress_vhosts_ssl_cert_failed",
        "Current number of ingresses with ssl true cert failed validation", nil, nil,)

    Status.ConfigErrors.desc = prometheus.NewDesc("ingress_errors_encountered",
        "Number of errors encountered during last ingress processing run", nil, nil,)

    Status.TemplateErrors.desc = prometheus.NewDesc("ingress_template_errors_encountered",
        "Number of errors rendering nginx.conf template", nil, nil,)

}

func (monitor *Monitor) Reset() {
    monitor.VHosts.counter = 0
    monitor.SslVHosts.counter = 0
    monitor.NonSslVHosts.counter = 0
    monitor.FailedVHosts.counter = 0
    monitor.FailedSslVHosts.counter = 0
    monitor.NoCertSslVHosts.counter = 0
    monitor.SslVHostsCertFail.counter = 0
    monitor.ConfigErrors.counter = 0
    monitor.TemplateErrors.counter = 0
}

func (monitor *Monitor) GetErrors() int {
    return int(monitor.ConfigErrors.counter+monitor.TemplateErrors.counter)
}

func (monitor *Monitor) IncVHosts() {
    monitor.VHosts.counter++
}

func (monitor *Monitor) IncSslVHosts() {
    monitor.SslVHosts.counter++
}

func (monitor *Monitor) IncNonSslVHosts() {
    monitor.NonSslVHosts.counter++
}

func (monitor *Monitor) IncFailedVHosts() {
    monitor.FailedVHosts.counter++
    monitor.ConfigErrors.counter++
}

func (monitor *Monitor) IncFailedSslVHosts() {
    monitor.FailedSslVHosts.counter++
    monitor.ConfigErrors.counter++
}

func (monitor *Monitor) IncNoCertSslVHosts() {
    monitor.NoCertSslVHosts.counter++
    monitor.ConfigErrors.counter++
}

func (monitor *Monitor) IncSslVHostsCertFail() {
    monitor.SslVHostsCertFail.counter++
    monitor.ConfigErrors.counter++
}

func (monitor *Monitor) IncTemplateErrors() {
    monitor.TemplateErrors.counter++
    monitor.ConfigErrors.counter++
}

func (monitor *Monitor) Describe(ch chan<- *prometheus.Desc) {
    ch <- monitor.VHosts.desc
    ch <- monitor.SslVHosts.desc
    ch <- monitor.NonSslVHosts.desc
    ch <- monitor.FailedVHosts.desc
    ch <- monitor.FailedSslVHosts.desc
    ch <- monitor.NoCertSslVHosts.desc
    ch <- monitor.SslVHostsCertFail.desc
    ch <- monitor.ConfigErrors.desc
    ch <- monitor.TemplateErrors.desc
}

func (monitor *Monitor) Collect(ch chan<- prometheus.Metric) {
    ch <- prometheus.MustNewConstMetric(monitor.VHosts.desc, prometheus.CounterValue, monitor.VHosts.counter)
    ch <- prometheus.MustNewConstMetric(monitor.SslVHosts.desc, prometheus.CounterValue, monitor.SslVHosts.counter)
    ch <- prometheus.MustNewConstMetric(monitor.NonSslVHosts.desc, prometheus.CounterValue, monitor.NonSslVHosts.counter)
    ch <- prometheus.MustNewConstMetric(monitor.FailedVHosts.desc, prometheus.CounterValue, monitor.FailedVHosts.counter)
    ch <- prometheus.MustNewConstMetric(monitor.FailedSslVHosts.desc, prometheus.CounterValue, monitor.FailedSslVHosts.counter)
    ch <- prometheus.MustNewConstMetric(monitor.NoCertSslVHosts.desc, prometheus.CounterValue, monitor.NoCertSslVHosts.counter)
    ch <- prometheus.MustNewConstMetric(monitor.SslVHostsCertFail.desc, prometheus.CounterValue, monitor.SslVHostsCertFail.counter)
    ch <- prometheus.MustNewConstMetric(monitor.ConfigErrors.desc, prometheus.CounterValue, monitor.ConfigErrors.counter)
    ch <- prometheus.MustNewConstMetric(monitor.TemplateErrors.desc, prometheus.CounterValue, monitor.TemplateErrors.counter)
}
