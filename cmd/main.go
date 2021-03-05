package main

import (
    "context"
    "errors"
    "fmt"
    v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "log"
    "net"
    "net/http"
    "os"
    "strings"
)

func main() {
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
    var namespace string
    if len(r.URL.Path[1:]) != 0 {
        namespace = r.URL.Path[1:]
    } else {
        namespace = "default"
    }
    fmt.Fprintf(w, "Hi there, Try to get resources from [%s] namespace.\n", namespace)

    client, err := k8sClient(w, r)
    if err != nil {
        fmt.Fprintf(w, "Something went wrong. I can't create a k8s client. [%s]\n", err)
    } else {
        fmt.Fprintln(w)
        writeConfigMaps(client, namespace, w)
        writeSectets(client, namespace, w)
        writePods(client, namespace, w)
    }
}

func writeConfigMaps(client *kubernetes.Clientset, namespace string, w http.ResponseWriter) {
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

func writeSectets(client *kubernetes.Clientset, namespace string, w http.ResponseWriter) {
    secrets, err := client.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{})
    if err != nil {
        fmt.Fprintf(w, "Something went wrong. I can't get the secrets. [%s]\n", err)
    } else {
        fmt.Fprintf(w, "Secrets\n========\n")
        for _, secret := range secrets.Items {
            fmt.Fprintf(w, " - %s\n", secret.Name)
        }
    }
}

func writePods(client *kubernetes.Clientset, namespace string, w http.ResponseWriter) {
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

func k8sClient(w http.ResponseWriter, r *http.Request) (*kubernetes.Clientset, error) {
    token, err := extractBearerToken(r)
    if err != nil {
        return nil, err
    }
    fmt.Fprintf(w, "Using authorization bearer token [%s]\n", token)

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
