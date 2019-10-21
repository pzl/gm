<template>
	<section id="dashboard">

		<div class="quickbar box">
			<div class="svc">
				Services: <span class="val">{{services}}</span>
			</div>
			<div class="vpn">
				VPN: <span class="val">{{vpnState}}</span>
			</div>
		</div>

		<div class="box">
			<table class="datatable">
				<tbody>
					<tr>
						<td>Linux</td>
						<td>{{linuxver}}</td>
					</tr>
					<tr>
						<td>rkt version</td>
						<td>{{rkt}}</td>
					</tr>
					<tr>
						<td>Podman Version</td>
						<td>{{podman}}</td>
					</tr>
					<tr>
						<td>Memory Used / Total</td>
						<td>{{kbSize(mem.used)}} / {{kbSize(mem.total)}}</td>
					</tr>
				</tbody>
			</table>
		</div>

	</section>
</template>

<script>
import axios from 'axios'

export default {
	data: function() {
		return {
			services: 0,
			vpn: false,
			linuxver: "",
			rkt: "",
			podman: "",
			mem: {
				total: 0,
				used: 0,
			}
		}
	},
	computed: {
		vpnState() { return this.vpn === true ? "up" : "down" },
	},
	components: {
	},
	methods: {
		kbSize: n => (n/2**20).toFixed(1)+" GB",
		get: url => axios.get(process.env.api+url),
	},
	mounted () {
		this.get("/api/services/count").then(c => {this.services = c.data })
		this.get("/api/system/versions").then(v => {
			this.linuxver = v.data.linux
			this.rkt = v.data.rkt
			this.podman = v.data.podman
		})
		this.get("/api/system/memory").then(m=> {
			this.mem.total = m.data.total
			this.mem.used = m.data.total - m.data.avail
		})

		this.get("/api/system/vpn").then(v=>{this.vpn = v.data})
	}
}
</script>

<style>

</style>
