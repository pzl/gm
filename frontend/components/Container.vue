<template>
	<div class="container">
		<h3>Container</h3>
		<div class="cid">ID: {{id}}</div>
		<div class="state">State: {{podstate}}</div>
		<div class="pid">PID: {{pid}}</div>
		<div class="times">
			<div class="created">Created: {{creation}}</div>
			<div class="started">Started: {{start}}</div>
		</div>
		<div>Exposed Ports: <span v-for="p in manifest.ports" :key="p.hostPort" class="host-port">{{p.hostPort}}➙{{p.name}}</span></div>
		<div>Host Mounts: <span v-for="v in manifest.volumes" :key="v.name" class="host-mount">{{v.source || "&lt;empty&gt;"}}➙{{v.name}}</span></div>
		<div class="net" v-if="networks">{{networks}}</div>
		<h3>Apps</h3>
		<div v-for="app in applist" class="app" :key="app.name">
			<h4 class="appname">{{app.name}}</h4>
			<div class="app-state">{{appstate(app)}}</div>
			<div class="appdef">
				<p>Run: <code>{{app.app.exec.join(" ")}}</code></p>
				<div class="mounts">
					<h5>Mounts</h5>
					<div v-for="m in app.app.mountPoints" :key="m.name">{{m.name}}: {{m.path}}</div>
				</div>
				<div class="ports" v-if="app.app.ports">
					<h5>Ports</h5>
					<div v-for="p in app.app.ports" :key="p.name">{{p.name}}: {{p.protocol}}:{{p.port}}</div>
				</div>
				<div class="env" v-if="app.app.environment">
					<h5>Environment</h5>
					<div v-for="e in app.app.environment" :key="e.name">{{e.name}}={{e.value}}</div>
				</div>
			</div>
			<div class="image">
				<h5>Image</h5>
				<div class="image-id">{{app.image.id}}</div>
				<div class="image-name">{{app.image.name}} @ {{app.image.version}}</div>
				<div class="labels"><span v-for="l in app.image.labels" :key="l.name">{{l.name}}: {{l.value}}</span></div>
			</div>
		</div>
	</div>
</template>


<script>
import merge from 'deepmerge'
const AppState = ["undefined", "running", "exited"]
const PodState = ["undefined", "embryo", "preparing", "prepared", "running", "aborted prepare", "exited", "deleting", "garbage"]

export default {
	props: {
		'id': {},
		'pid': {},
		'state': {},
		'apps': {},
		'manifest': {},
		'annotations': {},
		'cgroup': {},
		'created_at': {},
		'started_at': {},
		'networks': {},
	},
	computed: {
		podstate: function () {
			return PodState[this.state]
		},
		creation: function () {
			return new Date(this.created_at/1000000)
		},
		start: function () {
			return new Date(this.started_at/1000000)
		},
		applist: function () {
			let apps = []
			for (const a of this.apps) {
				let match = this.manifest.apps.find(x => a.image && x.image && a.image.id===x.image.id)
				if (match){
					apps.push(merge(a,match))
				}
			}
			return apps
		},
	},
	methods: {
		appstate:  (a) => AppState[a.state],
	}
}
</script>

<style>
.app {
	border: 1px solid black;
	padding: 20px;
	margin: 20px;
}

.host-mount, .host-port {
	display: block;
	margin-left: 50px;
}

</style>