<template>
	<section>
		<v-alert v-if="alert.show" :type="alert.type" dismissible>{{ alert.text }}</v-alert>

		<v-sheet>
			<div class="ram">
				<h2>Memory</h2>
				<p>Used: {{ memUsedGB }}</p>
				<p>Available: {{ memAvailGB }}</p>
				<p>Total: {{ memTotalGB }}</p>
			</div>
			
			<div class="disks">
				<h2>Disks</h2>
				<disk-diagram v-for="b in blocks" :key="b.blkid" v-bind="b" :max="maxBlockSize" />
			</div>

			<div class="versions">
				<h2>Version Info</h2>
				<p><strong>Linux</strong>: {{ versions.linux }}</p>
				<template v-if="versions.podman">
					<p><strong>Distro</strong>: {{ versions.podman.host.distribution.distribution }}</p>
					<p><strong>Hostname</strong>: {{ versions.podman.host.hostname }}</p>
					<p><strong>Uptime</strong>: {{ versions.podman.host.uptime }}</p>
					<p><strong>CPUs</strong>: {{ versions.podman.host.cpus }}</p>

					<p><strong>CGroup Version</strong>: {{ versions.podman.host.cgroupVersion }}</p>
					<p><strong>AppArmor</strong>: {{ versions.podman.host.security.apparmorEnabled }}</p>
					<p><strong>seccomp</strong>: {{ versions.podman.host.security.seccompEnabled }}</p>
					<p><strong>SELinux</strong>: {{ versions.podman.host.security.selinuxEnabled }}</p>
					<p><strong>Rootless</strong>: {{ versions.podman.host.security.rootless }}</p>
					<p><strong>Podman API</strong>: {{ versions.podman.version.APIVersion }}</p>
					<p><strong>Podman</strong>: {{ versions.podman.version.Version }}</p>
					<p><strong>Go</strong>: {{ versions.podman.version.GoVersion }}</p>
				</template>

			</div>


			<v-btn text color="orange" @click="sysDreload">Reload SystemD</v-btn>

		</v-sheet>
	</section>
</template>

<script>
import DiskDiagram from '~/components/DiskDiagram.vue'

const KBtoGb = n => (n/2**20).toFixed(1)+" GB"
const HumanSize = n => {
	const B = 1,
		  KB = 1024*B,
		  MB = 1024*KB,
		  GB = 1024*MB,
		  TB = 1024*GB
	if ( n > TB) {
		return (n/TB).toFixed(1)+" T"
	} else if (n > GB) {
		return (n/GB).toFixed(0)+" G"
	} else if (n > MB) {
		return (n/MB).toFixed(0)+" M"
	} else if (n > KB){
		return (n/KB).toFixed(0)+" K"
	} else {
		return n+" B"
	}
}
export default {
	data() {
		return {
			disks: [],
			memory: {
				avail: 0,
				total: 0,
			},
			versions: {
				linux: '',
				podman: null,
			},


			// alert banner
			alert: {
				text: '',
				show: false,
				type: '',
			}
		}
	},
	computed: {
		memUsed() { return this.memory.total - this.memory.avail },
		memUsedGB() { return KBtoGb(this.memUsed) },
		memAvailGB() { return KBtoGb(this.memory.avail) },
		memTotalGB() { return KBtoGb(this.memory.total) },
		blocks: function() {
			// Groups partitions into their top-level blocks

			// collect block names
			let blocks = this.disks.filter(d=>(d.Block.startsWith('sd') && d.Block.length == 3) || (d.Block.startsWith('nvme') && d.Block.length == 7)).map(d => d.Block)

			let blockmap = {}
			for (let b of blocks) {
				blockmap[b] = {
					disks: this.disks.filter(d => d.Block.startsWith(b) && d.Block !== b).sort((x,y)=>x.Block.localeCompare(y.Block)),
					blkid: b,
					rawsize: this.disks.find(d=>d.Block==b).RawSize,
				}
			}
			return blockmap
		},
		maxBlockSize: function() {
			return Math.max(...Object.values(this.blocks).map(v=>v.rawsize*1024))
		}
	},
	methods: {
		toHumanSize(n){ return HumanSize(n) },
		getMem() {
			this.$http.$get(`${this.$server}/api/v1/system/memory`)
				.then(d => {
					this.memory = d
				})
				.catch(e => {
					this.alert.type = 'error'
					const msg = 'response' in e ? e.response.data.error : e.message
					console.log(e)
					this.alert.text = `loading system memory failed: ${msg}`
					this.alert.show = true
				})
		},
		getDisks() {
			this.$http.$get(`${this.$server}/api/v1/system/disk`)
				.then(d => {
					this.disks = d
				})
				.catch(e => {
					this.alert.type = 'error'
					const msg = 'response' in e ? e.response.data.error : e.message
					console.log(e)
					this.alert.text = `loading system disks failed: ${msg}`
					this.alert.show = true
				})
		},
		getVersions() {
			this.$http.$get(`${this.$server}/api/v1/system/versions`)
				.then(d => {
					this.versions = d
				})
				.catch(e => {
					this.alert.type = 'error'
					const msg = 'response' in e ? e.response.data.error : e.message
					console.log(e)
					this.alert.text = `loading system versions failed: ${msg}`
					this.alert.show = true
				})
		},
		sysDreload() {
			this.$http.$post(`${this.$server}/api/v1/system/reload`)
				.then(d => {
					this.alert.type = "success"
					this.alert.text = "reloaded"
					this.alert.show = true
				})
				.catch(e => {
					this.alert.type = 'error'
					const msg = 'response' in e ? e.response.data.error : e.message
					console.log(e)
					this.alert.text = `reloading systemd failed: ${msg}`
					this.alert.show = true
				})
		}
	},
	mounted() {
		this.getMem()
		this.getDisks()
		this.getVersions()
	},
	beforeDestroy() {

	},
	components: { DiskDiagram },
}
</script>