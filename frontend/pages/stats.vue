<template>
	<section id="stats">
		<div class="box quickbar">
			<div>
				RAM <span class="val">{{kbSize(this.mem.avail)}} / {{kbSize(this.mem.total)}}</span>
			</div>
			<div>
				???
			</div>
		</div>

		<div class="box drive-overview">
			<div class="blocks">
				<diskblk v-for="b in blocks" :key="b.blkid" v-bind="b" :max="maxBlockSize"></diskblk>
			</div>
		</div>

		<HD v-for="s in disks.filter(d=>d.Block.length > 3)" :key="s.Block" v-bind="s" />
	</section>
</template>

<script>
import HD from '~/components/HD.vue'
import Diskblk from '~/components/Diskblk.vue'
import axios from 'axios'

export default {
	data: function () {
		return {
			disks: [],
			mem: {
				total: 0,
				avail: 0,
			}
		}
	},
	methods: {
		kbSize: n => (n/2**20).toFixed(1)+" GB"
	},
	computed: {
		blocks: function() {
			// collect just the top block devices, not partitions. E.g. sda, sdc  (NOT sdd1)
			let blocks = this.disks.filter(d=>d.Block.length==3).map(d => d.Block)

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
			//return Math.max(...Object.values(this.blocks).map(v=>v.disks.map(d=>d.All).reduce((acc,cur)=>acc+cur,0)),0)
		}
	},
	mounted () {
	  	Promise.all([
	  	    axios.get(process.env.api+'/api/stats/disk/'),
	  	    axios.get(process.env.api+"/api/system/memory/"),
	  	]).then(([disk, mem]) => {
	  		this.disks = disk.data.Disks
	  		this.mem.total = mem.data.total
	  		this.mem.avail = mem.data.avail
	  	})
	},

	components: { HD, Diskblk }
}
</script>


<style>
.drive-overview {
	padding: 25px;
}


@media screen and (min-width: 700px) {
	.blocks {
		width: 90%;
	}
}

</style>