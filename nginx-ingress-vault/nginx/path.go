package nginx

type Path struct {
    URI  string
    Service   string
    Namespace string
    Scheme    string
    Port      int32
}
