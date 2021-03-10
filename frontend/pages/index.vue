<template>
	<section>
		
		<div class="d-flex justify-space-around flex-column flex-sm-row flex-sm-wrap align-center align-content-center">
			<service
				v-for="(s,i) in services" :key="i" v-bind="s"
				@removed="services.splice(i,1)"
				@action="reloadService(s.name)"
				@toast="toast"
				class="my-7"
				:style="{ maxWidth: ($vuetify.mdAndUp ? '21%' : '100%') }"
			/>
		</div>

		<v-snackbar
			v-model="snacky.show"

		>
			{{ snacky.text }}
			<template v-slot:action="{ attrs }">
				<v-btn color="blue" text v-bind="attrs" @click="snacky.show = false">Close</v-btn>
			</template>
		</v-snackbar>

	</section>
</template>


<script>
import Service from '~/components/Service'

export default {
	data() {
		return {
			services: [],

			snacky: {
				show: false,
				text: ''
			}
		}
	},
	computed: {
	},
	methods: {
		getServices() {
			this.$http.$get(`${this.$server}/api/v1/services`)
				.then(d => {
					this.services = d
				})
				.catch(e => {
					this.toast(`Failed fetching services: ${e.message}`)
				})
		},
		reloadService(name) {
			this.$http.$get(`${this.$server}/api/v1/services/${name}`)
				.then(d => {
					const idx = this.services.map(s => s.name).indexOf(name)
					if (idx === -1) {
						this.services.unshift(d)
					} else {
						this.$set(this.services, idx, d)
					}
				})
				.catch(e => {
					console.log("reload fail: ",e)
					this.toast(`Reloading service ${name} failed: ${e.message}`)
				})
		},
		toast(msg) {
			this.snacky.text = msg
			this.snacky.show = true
		},
	},
	//async asyncData(context) {},
	mounted() {
		this.getServices()
	},
	beforeDestroy() {},
	watch: {},
	components: { Service }
}
</script>
