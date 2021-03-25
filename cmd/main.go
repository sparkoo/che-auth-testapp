package main

import (
    "context"
    "errors"
    "flag"
    "fmt"
    "html/template"
    "io"
    v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "log"
    "net"
    "net/http"
    "os"
    "strings"
)

var conf = &Conf{}

type Conf struct {
    Bind         string
    PodName      string
    PodNamespace string
}

type Page struct {
    Token                string
    Namespace            string
    Output               string
}

func (page *Page) Writeln(s string) {
    page.Write([]byte(s))
    page.Write([]byte("\n"))
}

func (page *Page) Write(p []byte) (n int, err error) {
    page.Output = page.Output + string(p)
    return len(p), nil
}


func main() {
    conf := parseArgs()
    http.HandleFunc("/", handler)
    http.HandleFunc("/query", handleQuery)
    log.Println("Listening on ", conf.Bind, " ...")
    log.Fatal(http.ListenAndServe(conf.Bind, nil))
}

func parseArgs() *Conf {
    flag.StringVar(&conf.Bind, "bind", ":8080", "server address to listen on")

    flag.Parse()

    conf.PodName = os.Getenv("POD_NAME")
    conf.PodNamespace = os.Getenv("POD_NAMESPACE")

    return conf
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
    if page, err := handle(w, r); err == nil {
        w.Write([]byte(page.Output))
    } else {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func handler(w http.ResponseWriter, r *http.Request) {
    if page, err := handle(w, r); err == nil {
        tmpl, err := template.ParseFiles("templates/index.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        if err := tmpl.Execute(w, page); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    } else {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func handle(w http.ResponseWriter, r *http.Request) (*Page, error) {
    log.Printf("%+v", r)

    page := &Page{}

    err := r.ParseForm()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return nil, err
    }

    // handle the namespace
    if namespace := r.Form.Get("namespace"); len(namespace) > 0 {
        log.Println("Using namespace from get param", namespace)
        page.Namespace = namespace
    } else {
        log.Println("using default namespace")
        page.Namespace = "che"
    }

    // handle token
    if token := r.Form.Get("token"); len(token) > 0 {
        page.Token = token
    } else {
        page.Token = ""
        if token, err := extractBearerToken(r); err == nil && len(token) > 0 {
            page.Token = token
            log.Println("Using bearer token from the authorization header.")
        }
    }
    log.Println("token: ", page.Token)

    page.Writeln("Hello, some info about me")
    page.Writeln("=========================")
    page.Writeln(fmt.Sprintf("Pod: [%s]", conf.PodName))
    page.Writeln(fmt.Sprintf("Namespace: [%s]", conf.PodNamespace))
    page.Writeln(fmt.Sprintf("Request path: [%s]", r.URL.Path))
    page.Writeln(fmt.Sprintf("Bind: [%s]", conf.Bind))
    page.Writeln("")
    page.Writeln("")

    page.Writeln(fmt.Sprintf("============ k8s request to [%s] namespace ============", page.Namespace))
    // request k8s
    client, err := k8sClient(page.Token)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return nil, err
    } else {
        writeConfigMaps(client, page.Namespace, page)
        writeSectets(client, page.Namespace, page)
        writePods(client, page.Namespace, page)
    }
    page.Writeln("========================================================")
    page.Writeln("")

    // authorization headers
    page.Writeln("Authorization Headers")
    page.Writeln("=====================")
    page.Writeln(strings.Join(r.Header["Authorization"], "\n"))
    page.Writeln("")

    // headers
    var sb strings.Builder
    // Loop over header names
    for name, values := range r.Header {
        // Loop over all values for the name.
        for _, value := range values {
            sb.WriteString(fmt.Sprintf("%s: %s\n", name, value))
        }
    }
    page.Writeln("Headers")
    page.Writeln("=============")
    page.Writeln(sb.String())
    page.Writeln("")

    return page, nil
}

func k8sClient(token string) (*kubernetes.Clientset, error) {
    config, err := k8sClientConfig(token)
    if err != nil {
        return nil, err
    }

    return kubernetes.NewForConfig(config)
}

func extractBearerToken(r *http.Request) (string, error) {
    authorizationHeader := r.Header.Get("Authorization")
    if len(authorizationHeader) == 0 {
        return "", errors.New("no Authorization header found")
    }

    if !strings.HasPrefix(authorizationHeader, "Bearer ") {
        return "", errors.New("no Authorization Bearer token found")
    }

    token := authorizationHeader[len("Bearer "):]

    return token, nil
}

func k8sClientConfig(token string) (*rest.Config, error) {
    host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
    if len(host) == 0 || len(port) == 0 {
        return nil, errors.New("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
    }

    var rootCaFile string
    if rootCaEnv, ok := os.LookupEnv("MINIKUBE_ROOT_CA"); ok {
        rootCaFile = rootCaEnv
    } else {
        rootCaFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
    }

    tlsClientConfig := rest.TLSClientConfig{
        CAFile: rootCaFile,
    }

    return &rest.Config{
        Host:            "https://" + net.JoinHostPort(host, port),
        TLSClientConfig: tlsClientConfig,
        BearerToken:     token,
    }, nil
}

func writeConfigMaps(client *kubernetes.Clientset, namespace string, w io.Writer) {
    configMaps, err := client.CoreV1().ConfigMaps(namespace).List(context.TODO(), v1.ListOptions{})
    if err != nil {
        fmt.Fprintf(w, "Something went wrong. I can't get the configMaps. [%s]\n", err)
    } else {
        fmt.Fprintf(w, "ConfigMaps\n========\n")
        for _, configMap := range configMaps.Items {
            fmt.Fprintf(w, " - %s\n", configMap.Name)
        }
    }
}

func writeSectets(client *kubernetes.Clientset, namespace string, w io.Writer) {
    secrets, err := client.CoreV1().Secrets(namespace).List(context.TODO(), v1.ListOptions{})
    if err != nil {
        fmt.Fprintf(w, "Something went wrong. I can't get the secrets. [%s]\n", err)
    } else {
        fmt.Fprintf(w, "Secrets\n========\n")
        for _, secret := range secrets.Items {
            fmt.Fprintf(w, " - %s\n", secret.Name)
        }
    }
}

func writePods(client *kubernetes.Clientset, namespace string, w io.Writer) {
    pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{})
    if err != nil {
        fmt.Fprintf(w, "Something went wrong. I can't get the pods. [%s]\n", err)
    } else {
        fmt.Fprintf(w, "Pods\n========\n")
        for _, pod := range pods.Items {
            fmt.Fprintf(w, " - %s\n", pod.Name)
        }
    }
}
