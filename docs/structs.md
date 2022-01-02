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
