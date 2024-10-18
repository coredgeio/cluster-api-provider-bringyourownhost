version = 2
[plugins]
{{- if ne .SANDBOX_IMAGE "" }}
  [plugins."io.containerd.grpc.v1.cri"]
    sandbox_image = "{{ .SANDBOX_IMAGE }}"
{{- end }}
  [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
    runtime_type = "io.containerd.runc.v2"
    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
      SystemdCgroup = true