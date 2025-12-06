local obs = obslua

local pocketstream = {
    process  = nil,
    disabled = false,
    token    = "",
    host     = "",
    source   = "",
    duration = 0.0,
}

function log_error(message)
    print("ERROR: " .. message)
    obs.blog(obs.LOG_ERROR, message)
end

function log_warn(message)
    print("WARN: " .. message)
    obs.blog(obs.LOG_WARNING, message)
end

function log_info(message)
    print("INFO: " .. message)
    obs.blog(obs.LOG_INFO, message)
end

function log_debug(message)
    print("DEBUG: " .. message)
    obs.blog(obs.LOG_DEBUG, message)
end

function script_load(settings)
    obs.obs_frontend_add_event_callback(handle_event)
end

function starts_with(message, prefix)
    local segment = string.sub(message, 1, string.len(prefix))
    return segment == prefix
end

function script_description()
    return "Helper plugin that auto-starts pocketstream executable."
end

function script_properties()
    local props = obs.obs_properties_create()

    obs.obs_properties_add_bool(props, "disabled", "Disable pocketstream autostart")
    obs.obs_properties_add_text(props, "token",   "Token",   obs.OBS_TEXT_DEFAULT)
    obs.obs_properties_add_text(props, "host",    "Host",    obs.OBS_TEXT_DEFAULT)
    obs.obs_properties_add_text(props, "source",  "Source",  obs.OBS_TEXT_DEFAULT)
    obs.obs_properties_add_float_slider(props, "duration", "Hls Chunk\nDuration", 0.0, 10.0, 0.1)

    return props
end

function script_update(settings)	
    pocketstream.disabled = obs.obs_data_get_bool(settings, "disabled")
    pocketstream.token    = obs.obs_data_get_string(settings, "token")
    pocketstream.host     = obs.obs_data_get_string(settings, "host")
    pocketstream.source   = obs.obs_data_get_string(settings, "source")
    pocketstream.duration = obs.obs_data_get_double(settings, "duration")
end

function handle_event(event)
    if pocketstream.disabled then
        log_info("Ignoring stream start. PocketStream script is set to 'disabled'.")
        return
    end

    -- log_debug("The event was = " .. event)

    if event == obs.OBS_FRONTEND_EVENT_STREAMING_STARTING then
        log_info("Starting PocketStream.")
        start_pocket_stream()
    elseif event == obs.OBS_FRONTEND_EVENT_STREAMING_STOPPED then
        pocketstream.process:close()
        log_info("Stream finished. PocketStream closed.")
    end
end

function start_pocket_stream()
    local token  = pocketstream.token
    if token == "" then
        log_error("Token cannot be empty. Please copy your token from the PocketWatch website.")
        return
    end

    local host = pocketstream.host
    if host == "" then
        log_error("Host cannot be empty. Please paste host url (for example: https://mydomain.example/")
        return
    end

    local args = " --dest " .. host .. " --token " .. token

    local source = pocketstream.source
    if source ~= "" then 
        args = args .. " --source " .. source
    end

    local duration = pocketstream.duration
    if duration ~= 0 then 
        args = args .. " --segment " .. duration
    end

    -- log_debug("Args are: " .. args)

    local process = io.popen("pocketstream" .. args)
    pocketstream.process = process;

    for line in process:lines() do
        if starts_with(line, "ERROR") then
            local message = string.sub(line, string.len("ERROR") + 2)
            log_error(message)
            return
        end

        if line == "PocketStream is ready" then
            break
        end
    end
end
