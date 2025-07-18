local obs = obslua

local pocketstream = {
    process  = nil,
    settings = {}
}

function script_load(settings)
    obs.obs_frontend_add_event_callback(handle_event)
end

function script_description()
    return "Helper plugin that auto-starts pocketstream executable."
end

function script_properties()
    local props = obs.obs_properties_create()

    obs.obs_properties_add_text(props, "token",  "Token",  obs.OBS_TEXT_DEFAULT)
    obs.obs_properties_add_text(props, "host",   "Host",   obs.OBS_TEXT_DEFAULT)
    obs.obs_properties_add_text(props, "source", "Source", obs.OBS_TEXT_DEFAULT)

    return props
end

function script_update(settings)	
	pocketstream.settings = settings
end

function handle_event(event)
    -- print("DEBUG: The event was = " .. event)

    if event == obs.OBS_FRONTEND_EVENT_STREAMING_STARTING then
        print("Starting PocketStream.")
        start_pocket_stream()
    elseif event == obs.OBS_FRONTEND_EVENT_STREAMING_STOPPED then
        pocketstream.process:close()
        print("Stream finished. PocketStream closed.")
    end
end

function start_pocket_stream()
    local token  = obs.obs_data_get_string(pocketstream.settings, "token")
    local host   = obs.obs_data_get_string(pocketstream.settings, "host")
    local source = obs.obs_data_get_string(pocketstream.settings, "source")

    local args = " --dest " .. host .. " --token " .. token

    if source ~= "" then 
        args = args .. " --source " .. source
    end

    -- print("DEBUG: Args are: " .. args)

    local process = io.popen("pocketstream" .. args, 'r')
    pocketstream.process = process;

    for line in process:lines() do
        if line == "PocketStream is ready" then
            break
        end
    end
end
