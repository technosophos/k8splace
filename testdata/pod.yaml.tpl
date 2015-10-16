apiVersion: v1
kind: Pod
metadata:
  name: deis-mc
  labels:
    heritage: deis
    version: 2015-sept
spec:
  restartPolicy: Never
  containers:
    - name: mc
      imagePullPolicy: {{ mc.pullPolicy }}
      image: deis/mc:2015-sept
      env:
        {% for key, value in mc.env %}
        {{key}}: {{value}}
        {% endfor %}
