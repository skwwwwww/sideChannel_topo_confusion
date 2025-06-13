function send_mirror_request(request_handle, headers, body)
    local cluster_name = "outbound|80||traffic-service-0.default.svc.cluster.local"
    headers:add("X-Traffic-Type", "confusion")
    local headers_map = {}
    for key, value in pairs(headers) do
    headers_map[key] = value
    end
    request_handle:logWarn("11110")
    local ok, err = request_handle:httpCall(
    cluster_name,
    headers_map,
    nil,
    5000
    )
    if not ok then
    request_handle:logWarn("Mirror request failed: " .. tostring(err))
    end
end

function envoy_on_request(request_handle)
    request_handle:logWarn("Intercepted request: skw666")
    local headers = request_handle:headers()
    local body = ""
    if request_handle:body() then
    body = request_handle:body():getBytes(0, request_handle:body():length())
    end
    send_mirror_request(request_handle, headers, body)
end