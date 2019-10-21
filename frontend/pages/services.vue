<template>
	<section id="services">
		<div class="box quickservices">
			<div v-for="s in services" :key="s.Name" class="svc-quick">
				<div class="status-bubble" :class="{ok: ok(s)}" :title="s.Name"></div>
			</div>
		</div>
		<service v-for="s in sortedServices" :key="s.Name" v-bind="s" />
	</section>
</template>

<script>
import Service from '~/components/Service.vue'
import axios from 'axios'

/* On the server-side, Manifest is a []byte, not a string. So when it gets
 * serialized, it goes from an object (struct) to base64 string. Changing it
 * there would require basically re-making the Pod struct already provided.
 *
 * OR, I can just expand it here for now.
 */
function expandManifests(data) {
	for (const srv of data) {
		if (srv.Container && srv.Runtime === 'rkt') {
			srv.Container.manifest = JSON.parse(atob(srv.Container.manifest))
		}
	}	

	return data
}


export default {
	data: function () {
		return {
			services: []
		}
	},
	computed: {
		sortedServices() {
			return this.services.slice().sort((a,b) => {
				if (a.LoadState !== b.LoadState) {
					return a.LoadState !== "loaded" ? -1 : 1
				}
				if (a.ActiveState !== b.ActiveState) {
					return a.ActiveState !== "active" ? -1 : 1 
				}
				if (a.SubState !== b.SubState) {
					return a.SubState !== "running" ? -1 : 1
				}
				return 0
			})
		}
	},
	mounted () {
		axios.get(process.env.api+'/api/services/')
			.then(response => this.services = expandManifests(response.data))
	},
	methods: {
		ok: function(s) {
			return s.LoadState == "loaded" && s.ActiveState == "active" && s.SubState == "running"
		},
	},
	components: { Service }
}
</script>


<style>

:root {
	--checkheight: 10px;
}

.quickservices {
	display: flex;
	justify-content: space-around;
	align-items: center;
}

.status-bubble {
	display: flex;
	align-items: center;
	justify-content: center;
	background-color: #E35321;
	border-radius: 100%;
	width: 20px;
	height: 20px;
}

.status-bubble.ok {
	background-color: transparent;
}

.status-bubble.ok::after {
	content: '';
	display: block;
	width: var(--checkheight);
	height: calc(var(--checkheight) * 2);
	border: solid #78BB00;
	border-width: 0 calc(var(--checkheight) * 0.8) calc(var(--checkheight) * 0.8) 0;
	transform: rotate(45deg);
	position: relative;
	top: calc(var(--checkheight) / 2 * -1);
}


</style>