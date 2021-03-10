<template>
	<v-badge overlap :content="dotContent" :icon="dotIcon" :color="dotColor" :value="showDot" :dot="smallDot">
		<v-card :loading="loading" class="service-card">
			<v-card-title><nuxt-link :to="`/services/${name}`">{{ name }}</nuxt-link></v-card-title>
			<v-card-subtitle>{{ description }}</v-card-subtitle>

			<v-card-text>
				Status: {{ status }}
			</v-card-text>

			<v-card-actions>
				<v-btn small v-if="ok" text color="red" @click="sendAction('stop')">Stop</v-btn>
				<v-btn small v-if="!ok" text color="green" @click="sendAction('start')">Start</v-btn>
				<v-btn small text color="orange" @click="sendAction('restart')">Restart</v-btn>
				<v-btn small v-if="enabled" color="red" text @click="sendAction('disable')">Disable</v-btn>
				<v-btn small v-else color="green" text @click="sendAction('enable')">Enable</v-btn>
				<v-btn small v-if="podman" text color="green" @click="sendAction('update')">Update</v-btn>
				<v-btn small icon color="red" title="Delete from watchlist. Does not affect underlying service" @click="unwatch"><v-icon>mdi-delete</v-icon></v-btn>
			</v-card-actions>
		</v-card>
	</v-badge>
</template>


<script>

export default {
	props: {
		name: {},
		description: {},
		path: {},
		active_state: {},
		load_state: {},
		sub_state: {},
		file_state: {},
		cmdline: {},
		followed: {},
		job_id: {},
		job_path: {},
		job_type: {},
		extended: {},
		external: {},
	},
	data() {
		return {
			loading: false,
		}
	},
	computed: {

		/*
		 systemctl --state=help

		  load_states:   stub, loaded, not-found, bad-setting, error, merged, masked
		  active_states: active, reloading, inactive, failed, activating, deactivating, maintenance
		  sub_states:    dead, waiting, running, failed 
(service) sub_states:    dead condition start(-(pre|post))? running, exited, reload, stop(-(watchdog|post))? stop-sig(term|kill) final-watchdog final-sig(term|kill) failed auto-restart cleaning

		*/
		smallDot() { return false },
		showDot() { return true },
		dotIcon() { return this.loading ? 'mdi-timer-sand' : '' },
		dotContent() { return '' },
		dotColor() {
			if (this.loading) {
				return 'purple'
			}

			if (["not-found", "bad-setting","error"].indexOf(this.load_state) !== -1) {
				return 'red'
			}
			if (["failed","bad","deactivating"].indexOf(this.active_state) !== -1) {
				return 'red'
			}
			if (["dead","failed","exited","stop","stop-watchdog","stop-post","stop-sigterm","stop-sigkill"].indexOf(this.sub_state) !== -1) {
				return 'red'
			}

			if (this.active_state == "inactive") {
				return 'grey'
			}

			if (["start-pre","start-post","running","reload","listening"].indexOf(this.sub_state) !== -1) {
				return 'green'
			}

			return 'orange'
		},
		status() {
			switch (this.load_state) {
				case "not-found":
				case "bad-setting":
					return this.load_state
			}

			switch (this.active_state) {
				case "inactive":
				case "failed":
				case "activating":
				case "deactivating":
					return this.active_state
			}

			return this.sub_state

		},
		ok() {
			return this.load_state == "loaded" && this.active_state == "active" && (this.sub_state == "running" || this.sub_state == "listening")
		},
		enabled() {
			return this.file_state == "enabled"
		},
		podman() {
			return ('runtime' in this.extended && this.extended.runtime == "podman")
		}
	},
	methods: {
		unwatch() {
			this.loading = true
			this.$http.$delete(`${this.$server}/api/v1/services/${this.name}`)
			.then(d => {
				this.loading = false
				this.$emit('toast',`${this.name} removed`)
				this.$emit('removed')
			})
			.catch(e => {
				this.loading = false
				this.$emit('toast', `Removal failed: ${e.message}`)
			})
		},
		sendAction(actn) {
			this.$http.$post(`${this.$server}/api/v1/services/${this.name}`, { action: actn })
				.then(d => {
					this.$emit('action',actn)
					this.loading = false
					this.makeActionToast(actn, d)

				})
				.catch(e => {
					this.loading = false
					const msg = 'response' in e ? e.response.data.error : e.message
					this.$emit('toast', `${actn} ${this.name} failed: ${msg}`)
					console.log(`Action '${actn}' ${this.name} failed: `,JSON.stringify(e))
				})
		},
		makeActionToast(actn, resp) {
			const toast = (m) => { this.$emit('toast', m) }
			switch (actn) {
				case 'disable': return toast(resp.changes.map(c => `${c.Type}ed ${c.Filename}`).join(';'))
				case 'enable':  return toast(resp.changes.map(c => `${c.Type}ed ${c.Filename} -> ${c.Destination}`).join(';'))
				case 'start':
				case 'stop':
				case 'restart': return toast(`${actn}ed ${this.name}`)

			}
		}
	},
	mounted(){

	},
	beforeDestroy() {

	},
	components: {},
}
</script>

<style>
.service-card .v-card__title a {
	text-decoration: none;
	color: black;
}
</style>