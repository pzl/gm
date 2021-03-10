<template>
	<v-app dark>
		<v-app-bar app flat color="blue darken-2" class="white--text">
			<v-container>
				<div class="d-flex align-center">
					<div class="d-flex justify-start align-center mr-10">
						<img src="/icon.png" height="30" width="30" class="mr-3" />
						<v-toolbar-title class="text-h4 font-weight-black"><nuxt-link to="/"><span class="d-none d-md-inline">mr manager</span><span class="d-inline d-md-none">gm</span></nuxt-link></v-toolbar-title>
					</div>
					<div cols="auto">
						<v-btn nuxt to="/services/vpn-out.service" icon :title="vpnState"><v-icon :color="vpnColor">{{ vpnIcon }}</v-icon></v-btn>
					</div>
					<v-spacer />
					<nuxt-link to="/system">System</nuxt-link>
					<v-spacer />
					<div cols="auto">
						<v-menu v-model="addPanel.show" :close-on-content-click="false" :nudge-width="200" offset-y>
							<template v-slot:activator="{ on, attrs }">
								<v-btn color="green darken-1 white--text" rounded v-on="on" v-bind="attrs">
									<v-icon :left="$vuetify.mdAndUp">mdi-plus-thick</v-icon><span class="d-none d-md-inline">Add Service</span>
								</v-btn>
							</template>
							<v-card>
								<v-card-text>
									<v-autocomplete
										v-model="addPanel.value"
										:items="addPanel.items"
										:loading="addPanel.loading"
										:search-input.sync="addPanel.search"
										:cache-items="true"
										:clearable="true"
										hide-details
										hide-no-data
									>
									</v-autocomplete>
								</v-card-text>

								<v-card-actions>
									<v-btn @click="addService">Add</v-btn>
								</v-card-actions>
							</v-card>
						</v-menu>
					</div>
				</div>
			</v-container>
		</v-app-bar>
		<v-main>
			<v-container>
				<nuxt />
			</v-container>
		</v-main>
		<v-footer app>
			<span>&copy; {{ new Date().getFullYear() }}</span>
		</v-footer>
	</v-app>
</template>

<script>
export default {
	name: 'default',
	data() {
		return {
			vpnState: 'success',

			addPanel: {
				show: false,
				value: '',
				search: null,
				items: [],
				loading: false,
			},
		}
	},
	computed: {
		vpnIcon() {
			switch (this.vpnState) {
				case 'running': return 'mdi-shield-lock'
				case 'dead': return 'mdi-shield-remove-outline'
				case 'unknown': // fallthrough
				default: return 'mdi-shield-alert'
			}
		},
		vpnColor() {
			switch (this.vpnState) {
				case 'running': return 'success';
				case 'unknown': //fallthrough
				default: return 'orange darken-1'
			}
		}
	},
	methods: {
		setTheme() {
			if (window && window.matchMedia && window.matchMedia('(prefers-color-scheme:dark').matches) {
				this.$vuetify.theme.dark = true;
			} else {
				this.$vuetify.theme.dark = false
			}
		},
		getVPNStatus() {
			this.$http.$get(`${this.$server}/api/v1/services/vpn-out.service`)
				.then(d => {
					this.vpnState = d.sub_state
				})
				.catch(e => {
					this.vpnState = 'error'
				})
		},
		addService() {
			this.$http.$post(`${this.$server}/api/v1/services`, { name: this.addPanel.value })
			.then(d => {
				// @todo: notify page of new service?
				// or force a refresh?
				this.addPanel.show = false
				this.addPanel.value = ''
			})
			.catch(e => {
				// show the err somewhere
			})
		}
	},
	watch: {
		'addPanel.search'(val) {
			this.addPanel.loading = true
			this.$http.$get(`${this.$server}/api/v1/services/search?q=${val}`)
				.then(d => {
					this.addPanel.items = d
					this.addPanel.loading = false
				})
				.catch(e => {
					this.addPanel.loading = false
				})
		}
	},
	mounted() {
		this.getVPNStatus()
		this.setTheme()
		if (window && window.matchMedia) {
			window.matchMedia('(prefers-color-scheme:dark)').addListener(e => {
				if (e.matches) {
					this.$vuetify.theme.dark = true
				} else {
					this.$vuetify.theme.dark = false
				}
			})
		}
	}
}
</script>


<style>
html {
	font-size: 16px;
	word-spacing: 1px;
	-ms-text-size-adjust: 100%;
	-webkit-text-size-adjust: 100%;
	-moz-osx-font-smoothing: grayscale;
	-webkit-font-smoothing: antialised;
	box-sizing: border-box;
}

*, *::before, *::after {
	box-sizing: border-box;
	margin: 0;
}


.v-application .v-toolbar__content a {
	color: white;
	text-decoration: none;
}
</style>
