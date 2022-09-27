worker_processes auto;
error_log {{ .NginxWorkDir | ConfirmSlash }}error.log;
pid {{ .NginxWorkDir | ConfirmSlash }}nginx.pid;

worker_rlimit_nofile 65535;
events {
    use epoll;
    worker_connections 65535;
}

http {
    {{ if .EnableWebSocketProxy }}
    map $http_upgrade $connection_upgrade {
        default upgrade;
        '' close;
    }
    {{ end }}
    log_format main '{'
        '"record_time": "$time_iso8601",'
        '"status": "$status",'
        '"remote_addr": "$remote_addr",'
        '"http_x_forwarded_for": "$http_x_forwarded_for",'
        '"upstream_addr": "$upstream_addr",'
        '"host": "$http_host"'
        '"request_uri": "$request_uri",'
        '"request_method": "$request_method",'
        '"http_user_agent": "$http_user_agent",'
        '"http_referer": "$http_referer",'
        '"request_time": "$request_time",'
        '"upstream_response_time": "$upstream_response_time",'
        '"body_bytes_sent": "$body_bytes_sent",'
    '}';
    access_log {{ .NginxWorkDir | ConfirmSlash }}access.log main;

    server {
        listen 443 ssl {{ if .EnableListenHTTP2 }}http2{{ end }};
        server_name {{ .Server }};
        ssl_certificate {{ .SSLCertificatePath }};
        ssl_certificate_key {{ .SSLPrivateKeyPath }};
        {{ if .EnableModernTLSOnly }}
        ssl_protocols TLSv1.3;
        ssl_prefer_server_ciphers off;
        {{ else }}
        ssl_protocols TLSv1 TLSv1.1 TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384:DHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA:ECDHE-RSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES256-SHA256:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:DES-CBC3-SHA;
        ssl_prefer_server_ciphers on;
        {{ end }}
        ssl_session_timeout 5m;
        client_header_buffer_size 16k;
        location ^~/{{ .ComplexPath }}/ {
            resolver 1.1.1.1 ipv6=off;
            {{range $key,$value := .CustomHeaders}}
            if ( $http_{{ $key | ToLower }} != '{{ $value }}' ){
                return 404;
            }
            {{end}}
            set $_full_uri $uri$is_args$args;
            if ( $_full_uri ~ /{{ .ComplexPath }}/([^/]+)/(.*) ){
                set $_host $1;
                set $_uri $2;
            }
            proxy_pass $scheme://$_host/$_uri;
            proxy_redirect https://{{ .Server }}/{{ .ComplexPath }}/ /;
            proxy_buffer_size 256k;
            proxy_buffers 64 32k;
            proxy_busy_buffers_size 1m;
            proxy_temp_file_write_size 512k;
            proxy_max_temp_file_size 128m;
            proxy_set_header Host $_host;
            proxy_ssl_server_name on;
            {{ range $key,$value := .CustomHeaders}}
            proxy_set_header {{ $key | ToLower }} '';
            {{end}}
            {{ if .EnableWebSocketProxy }}
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
            {{ end }}
        }
        location / {
           resolver 1.1.1.1;
           return 404;
        }
    }
}
