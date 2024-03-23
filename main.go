package main


import (
    "context"
    "io"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
    "log"
    "fmt"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/mount" // Added import for mount
    "os"
    "path/filepath"
    "time"
    "archive/zip"
)

// Helper function to copy data between streams
func ioCopy(dst io.Writer, src io.Reader) {
    _, err := io.Copy(dst, src)
    if err != nil {
        log.Fatal(err)
    }
}

func buildZip(src io.Reader) {
    zipFile, err := os.Create("mas-must-gather.zip")
    if err != nil {
        log.Fatal(err)
    }
    defer zipFile.Close();
    zipWriter := zip.NewWriter(zipFile);
    defer zipWriter.Close()
    outputFile, err := zipWriter.Create("must-gather.log")
    if err != nil {
        log.Fatal(err)
    }
    _, err = io.Copy(outputFile, src) 
    if err != nil {
        log.Fatal(err);
    }
}

func runCommand(cli *client.Client, containerId string, command string) (error) {
    execConfig := types.ExecConfig{
        AttachStdout: true,
        AttachStderr: true,
        Tty: true, 
        Cmd:          []string{"sh", "-c", command},
    }

    execResp, err := cli.ContainerExecCreate(context.Background(), containerId, execConfig)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(execResp)
    time.Sleep(time.Second * 1);

    execId := execResp.ID;
    startResp, err := cli.ContainerExecAttach(context.Background(), execId, types.ExecStartCheck{})
    if err != nil {
        log.Fatal(err)
    }
    defer startResp.Close()

    for {
        output := make([]byte, 4096)
        _, err = startResp.Reader.Read(output)
        if err != nil {
            break;
        }
        fmt.Println("Exec command output:", string(output))
    }

    return nil;
}


func main() {

    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go 'oc login ...'")
        os.Exit(1)
    }

    ocLoginCommand := os.Args[1]

    IMAGE_NAME := "quay.io/ibmmas/cli"

    // Create a new Docker client
    cli, err := client.NewClientWithOpts(client.FromEnv)
    if err != nil {
        log.Fatal(err)
    }

    // Pull the image if not already present
    reader, err := cli.ImagePull(context.Background(), IMAGE_NAME, types.ImagePullOptions{})
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()
    ioCopy(os.Stdout, reader)

    // Define container configuration
    config := &container.Config{
        // Image: "quay.io/ibmmas/cli",
        Image: IMAGE_NAME,
        StopTimeout: &[]int{10}[0],
        Tty: true, 
        AttachStdin: false, 
        AttachStdout: true, 
        AttachStderr: true,
    }

    // Get the current working directory
    absPath, err := filepath.Abs(".")
    if err != nil {
        log.Fatal(err)
    }

    // Define host configuration
    hostConfig := &container.HostConfig{
        // Mount the home directory
        Mounts: []mount.Mount{
            {
                Type:   mount.TypeBind,
                Source: absPath,
                Target: "/mnt/home",
            },
        },
    }

    // Create the container
    resp, err := cli.ContainerCreate(context.Background(), config, hostConfig, nil, nil, "")
    if err != nil {
        log.Fatal(err)
    }

    // Start the container
    cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{});

    // cmd := "oc login --token=sha256~U4jzm4e0_23X_T3VrVUUdDjwb0K-PDbT42quoIDrM-s --server=https://api.sms-oc4-youbim.sms-inc.net:6443 --insecure-skip-tls-verify"
    cmd := ocLoginCommand +  " " +  "--insecure-skip-tls-verify"
    runCommand(cli, resp.ID, cmd);

    mustGather := "mas must-gather"
    runCommand(cli, resp.ID, mustGather);

    // Copy files from the container to the host
    rc, _, err := cli.CopyFromContainer(context.Background(), resp.ID, "/tmp/must-gather")
    if err != nil {
        log.Fatal(err)
    }
    defer rc.Close()
    buildZip(rc)
    fmt.Println("Files copied successfully")

    // Stop and remove the container
    err = cli.ContainerStop(context.Background(), resp.ID, container.StopOptions{})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Container stopped")

    err = cli.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{Force: true}) // Add Force: true to remove running container
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Container removed")
    fmt.Println("")
    fmt.Println("")
    fmt.Println("-----")
    fmt.Println("Done.")
    fmt.Println("-----")
    fmt.Println("")
    fmt.Println("")

}
