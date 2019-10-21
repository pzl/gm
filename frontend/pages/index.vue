<template>
  <section id="dashboard">

  	<div class="quickbar box">
  		<div class="svc">
  			Services: <span class="val">{{services}}</span>
  		</div>
  		<div class="vpn">
  			VPN: <span class="val">{{vpn}}</span>
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
	  				<td>Memory</td>
	  				<td>{{kbSize(mem.total)}}</td>
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
  		vpn: "down",
  		linuxver: "",
  		rkt: "",
      podman: "",
  		mem: {
  			total: 0,
  		}
  	}
  },
  components: {
  },
  methods: {
  	kbSize: n => (n/2**20).toFixed(1)+" GB",
    get: url => axios.get(process.env.api+url),
  },
  mounted () {
  	Promise.all([
  	    axios.get(process.env.api+"/api/services/count/"),
  	    axios.get(process.env.api+"/api/system/versions/"),
  	    axios.get(process.env.api+"/api/system/memory/"),
  	    axios.get(process.env.api+"/api/system/vpn/"),
  	]).then(([count,vers, mem, vpn]) => {
  		this.services = count.data
  		this.linuxver = vers.data.linux
  		this.rkt = vers.data.rkt
  		this.podman = vers.data.podman
  		this.mem.total = mem.data.total
  		this.vpn = vpn.data === true ? "up" : "down"
  	})
  }
}
</script>

<style>

</style>
