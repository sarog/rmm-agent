#### AgentUpdate

```python
# api/tacticalrmm/agents/tasks.py:57
nats_data = {
    "func": "agentupdate",
    "payload": {
        "url": url,
        "version": version,
        "inno": inno
    }
}
```

#### GetEventLog

```python
# api/tacticalrmm/agents/views.py:309
data = {
    "func": "eventlog",
    "timeout": timeout,
    "payload": {
        "logname": logtype,
        "days": str(days),
    },
}
```

#### GetSoftware, InstallWithChoco

```python
# api/tacticalrmm/software/views.py:56
nats_data = {
    "func": "installwithchoco",
    "choco_prog_name": name,
    "pending_action_pk": action.pk,
}
```

#### RunScript, RunScriptFull

```python
# api/tacticalrmm/agents/models.py:338
data = {
    "func": "runscriptfull" if full else "runscript",
    "timeout": timeout,
    "script_args": parsed_args,
    "payload": {
        "code": script.code,
        "shell": script.shell
    }
}
```

#### SendRawCMD

```python
# api/tacticalrmm/agents/views.py:329
data = {
    "func": "rawcmd",
    "timeout": timeout,
    "payload": {
        "command": request.data["cmd"],
        "shell": request.data["shell"],
    },
}
```
