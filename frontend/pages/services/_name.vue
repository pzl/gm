<template>
	<div class="thing-page">
		<v-alert v-if="alert.show" :type="alert.type">{{ alert.text }}</v-alert>

		<h1>{{ name }}</h1>
		<h2 class="text-subtitle-1">{{ description }}</h2>

		<p><strong>Active State</strong>: {{ active_state }}</p>
		<p><strong>Load State</strong>: {{ load_state }}</p>
		<p><strong>Sub State</strong>: {{ sub_state }}</p>
		<p><strong>File State</strong>: {{ file_state }}</p>

		<p v-if="cmdline"><strong>Command line</strong>: <code>{{ cmdline.join(' ') }}</code></p>

		<p><strong>PID</strong> {{ extended.pid }}</p>
		<p><strong>Restarts</strong> {{ extended.restarts }}</p>
		<p><strong>Memory</strong> {{ extended.mem }}</p>
		<p><strong>Last Change</strong> {{ last }}</p>
		<p><strong>Runtime</strong> {{ extended.runtime }}</p>

		<div v-if="external">
			<p><strong>Name</strong>: {{ external.Name }}</p>
			<p><strong>Container ID</strong>: {{ external.Id }}</p>
			<p><string>Privileged</string>: {{ external.HostConfig.Privileged }}</p>
			<v-list v-if="external.HostConfig.SecurityOpt && external.HostConfig.SecurityOpt.length" subheader two-line>
				<v-subheader inset>SecurityOpt</v-subheader>
				<v-list-item v-for="(e,i) in external.HostConfig.SecurityOpt" :key="`secopt-${i}`">
					<v-list-item-avatar><v-icon class="grey lighten-1" dark>mdi-console-network</v-icon></v-list-item-avatar>
					<v-list-item-content>
						<v-list-item-title>{{ e.split('=').pop() }}</v-list-item-title>
						<v-list-item-subtitle>{{ e.split('=')[0] }}</v-list-item-subtitle>
					</v-list-item-content>
				</v-list-item>
			</v-list>
			<p><strong>Args</strong>: {{ external.Args.join(' ') }}</p>
			<p><strong>EntryPoint</strong>: {{ external.Config.EntryPoint }}</p>
			<p><strong>Container State</strong>: {{ external.State.Status }}</p>
			<p><strong>Image</strong>: {{ external.ImageName }}</p>
			<p v-if="external.Pod"><strong>Pod</strong>: {{ external.Pod }}</p>
			<v-list v-if="external.Config.Env && external.Config.Env.length" subheader two-line>
				<v-subheader inset>Env</v-subheader>
				<v-list-item v-for="(e,i) in external.Config.Env" :key="`env-${i}`">
					<v-list-item-icon><v-icon color="grey" dark>mdi-console-network</v-icon></v-list-item-icon>
					<v-list-item-content>
						<v-list-item-title>{{ e.split('=').pop() }}</v-list-item-title>
						<v-list-item-subtitle>{{ e.split('=')[0] }}</v-list-item-subtitle>
					</v-list-item-content>
				</v-list-item>
			</v-list>
			<v-list v-if="external.Mounts && external.Mounts.length" subheader>
				<v-subheader inset>Mounts</v-subheader>
				<v-list-item v-for="(m,i) in external.Mounts" :key="`mount-${i}`">
					<v-list-item-icon><v-icon color="grey" dark>mdi-folder-move</v-icon></v-list-item-icon>
					<v-list-item-content><v-list-item-title>{{ m.Source }} -> {{ m.Destination }}</v-list-item-title><v-icon v-if="!m.RW" title="Read Only">mdi-file-cancel</v-icon></v-list-item-content>
				</v-list-item>
			</v-list>
			<v-list v-if="external.NetworkSettings.Ports && Object.keys(external.NetworkSettings.Ports).length" subheader>
				<v-subheader inset>Ports</v-subheader>
				<v-list-item v-for="(dest,src) in external.NetworkSettings.Ports" :key="`port-${src}`">
					<v-list-item-icon><v-icon color="grey" dark>mdi-lan-connect</v-icon></v-list-item-icon>
					<v-list-item-content><v-list-item-title>{{ src }} -> {{ dest[0].HostPort }}</v-list-item-title></v-list-item-content>
				</v-list-item>
			</v-list>
			<v-list v-if="external.HostConfig.Devices && external.HostConfig.Devices.length" subheader>
				<v-subheader inset>Devices</v-subheader>
				<v-list-item v-for="(d,i) in external.HostConfig.Devices" :key="`device-${i}`">
					<v-list-item-icon><v-icon color="grey" dark>mdi-expansion-card</v-icon></v-list-item-icon>
					<v-list-item-content><v-list-item-title>{{ d.PathInContainer }} -> {{ d.PathOnHost }}</v-list-item-title></v-list-item-content>
				</v-list-item>
			</v-list>
			<v-list v-if="external.Config.Labels && Object.keys(external.Config.Labels).length" subheader two-line>
				<v-subheader inset>Labels</v-subheader>
				<v-list-item v-for="(val,key) in external.Config.Labels" :key="`label-${key}`">
					<v-list-item-icon><v-icon color="grey" dark>mdi-tag</v-icon></v-list-item-icon>
					<v-list-item-content>
						<v-list-item-title>{{ val }}</v-list-item-title>
						<v-list-item-subtitle>{{ key }}</v-list-item-subtitle>
					</v-list-item-content>
				</v-list-item>
			</v-list>
			<v-list v-if="external.Config.Annotations && Object.keys(external.Config.Annotations).length" subheader two-line>
				<v-subheader inset>Annotations</v-subheader>
				<v-list-item v-for="(val,key) in external.Config.Annotations" :key="`annot-${key}`">
					<v-list-item-icon><v-icon color="grey" dark>mdi-tag</v-icon></v-list-item-icon>
					<v-list-item-content>
						<v-list-item-title>{{ val }}</v-list-item-title>
						<v-list-item-subtitle>{{ key }}</v-list-item-subtitle>
					</v-list-item-content>
				</v-list-item>
			</v-list>
			<v-btn icon @click="showDetails = !showDetails"><v-icon left>mdi-chevron-{{ showDetails ? 'up' : 'down'}}</v-icon> Expand</v-btn>
			<v-expand-transition>
				<div v-show="showDetails">
					<pre>{{ external }}</pre>
				</div>
			</v-expand-transition>
		</div>

		<div class="actions">
			<v-btn v-if="running" text color="red" @click="sendAction('stop')">Stop</v-btn>
			<v-btn v-else text color="green" @click="sendAction('start')">Start</v-btn>
			<v-btn text color="orange" @click="sendAction('restart')">Restart</v-btn>
			<v-btn v-if="enabled" color="red" text @click="sendAction('disable')">Disable</v-btn>
			<v-btn v-else color="green" text @click="sendAction('enable')">Enable</v-btn>
			<v-btn v-if="extended.runtime == 'podman'" text color="green" @click="sendAction('update')">Update</v-btn>
			<v-btn icon color="red" title="Delete from watchlist. Does not affect underlying service" @click="unwatch"><v-icon>mdi-delete</v-icon></v-btn>
		</div>

	</div>
</template>


<script>
import { format } from 'date-fns';


export default {
	data () {
		return {
			name: '',
			description: '',
			path: '',
			active_state: '',
			load_state: '',
			sub_state: '',
			cmdline: [],
			followed: '',
			job_id: '',
			job_path: '',
			job_type: '',
			file_state: '',
			extended: {
				pid: -1,
				restarts: -1,
				mem: -1,
				last_change: 0,
				runtime: "unknown",
			},
			external: null,


			//locals
			loading: false,
			showDetails: false,
			// alert banner
			alert: {
				text: '',
				show: false,
				type: '',
			}
		}
	},
	computed: {
		last() {
			return this.extended.last_change ? format(new Date(this.extended.last_change/1000), 'EE PPpp '): ''
		},
		running() {
			return this.load_state == "loaded" && (this.active_state == "active" || this.active_state == "activating") && (this.sub_state == "running" || this.sub_state == "auto-restart" || this.sub_state == "listening")
		},
		enabled() {
			return this.file_state == "enabled"
		}
	},
	async asyncData(context) {
		const data =  await context.$http.$get(`${context.$server}/api/v1/services/${context.params.name}`)
							.catch(e => {
								console.log(e)
								context.error(e)
							})
		return data
	},
	methods: {
		async unwatch() {
			this.loading = true;
			this.alert.show = false;
			await this.$http.$delete(`${this.$server}/api/v1/services/${this.name}`)
				.then(() => {
					this.$router.push({
						path: "/"
					})
				})
				.catch(e => {
					this.loading = false;
					console.log(e)
					this.alert.text = e.message;
					this.alert.type = "error"
					this.alert.show = true;
				})
		},
		async sendAction(actn) {
			this.alert.show = false;
			this.loading = true;

			await this.$http.$post(`${this.$server}/api/v1/services/${this.name}`, { action: actn })
				.then(d => {
					this.loading = false
					this.$nuxt.refresh()
				})
				.catch(e => {
					this.loading = false
					this.alert.type = "error"
					const msg = 'response' in e ? e.response.data.error : e.message
					this.alert.text = `${actn} ${this.name} failed: ${msg}`
					this.alert.show = true
				})
		},
	},
	watch: {},
	components: {}
}
</script>


<style>

</style>
