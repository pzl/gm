<template>
	<div class="disk-diagram d-sm-flex justify-space-between align-center">
		<div class="flex-grow-1">
			<div class="diskblk d-flex" :style="{ width: width }">
				<div v-for="d in disks" :key="d.FS" :style="diskStyle(d)" class="partition d-flex align-center justify-space-around">
					<div class="mount">{{d.Mount.replace("/run/media/dan/","")}}</div>
					<div class="perc" v-if="d.Mount">{{diskUsedPerc(d)}}%</div>
				</div>
			</div>
		</div>
		<p class="blkid ml-4">{{blkid}}</p>
	</div>
</template>


<script>

export default {
	props: {
		blkid: {},
		disks: {},
		rawsize: {},
		max: {}
	},
	data() {
		return {

		}
	},
	computed: {
		blkSize() { return this.rawsize*1024 },
		width() { return  this.blkSize / this.max * 100 + '%' },
	},
	methods: {
		diskSize(d){ return  d.RawSize * 1024 },
		diskWidth(d) { return this.diskSize(d) / this.blkSize * 100; },
		diskUsedPerc(d) {
			return d.Mount === "" ? -1 : Math.round(d.Used / this.diskSize(d) * 100);
		},
		diskStyle: function(d) {
			let width=this.diskWidth(d),
				used = this.diskUsedPerc(d),
				color;

			
			if (used > 85) {
				color = '#FFC0CB'
			} else if (used > 60) {
				color = '#F3D4A4'
			} else if (used > 40) {
				color = '#E6E48E'
			} else if (used === -1) {
				color = "#efefef"
				used=width
			} else {
				color = '#81BA85'
			}

			return {
				width: width+'%',
				background: 'linear-gradient(90deg, '+color+' '+used+'%, #FFFFFF '+used+'%)'
			}
		}
	},
	components: {},
}
</script>

<style>
.diskblk {
	border: 1px solid #777;
	margin: 10px 0;
	flex-shrink: 0;
	flex-grow: 0;
	min-height: 2rem;
}

.partition {
	border-left: 1px solid #8d8d8d;
	position: relative;
	font-size: 13px;
}
.partition:first-child {
	border-left: none;
	min-height: 30px;
}

.partition .mount {
	font-size: 12px;
	position: absolute;
	top: 2px;
	left: 2px;
}


.partition .perc {
	margin-top: 13px;
	color: rgba(0, 0, 0, 0.4);
}

</style>