<template>
	<div class="blk">
		<div class="diskblk" :style="{ width: width }">
			<div v-for="d in disks" :key="d.FS" :style="diskStyle(d)" class="partition">
				<div class="mount">{{d.Mount.replace("/run/media/dan/","")}}</div>
				<div class="perc" v-if="d.Mount">{{diskUsedPerc(d)}}%</div>
			</div>
		</div>
		<p class="blkid">{{blkid}}</p>
	</div>
</template>


<script>


export default {
	props: {
		'blkid': {},
		'disks': {},
		'rawsize': {},
		'max': {}
	},
	computed: {
		blkSize: function () {
			return this.rawsize*1024
			//return this.disks.map(d=>d.All).reduce((acc,cur)=>acc+cur,0)
		},
		width: function () {
			return  this.blkSize / this.max * 100 + '%'
		}
	},
	methods: {
		diskSize: function (d) {
			return d.All ? d.All : d.RawSize * 1024
		},
		diskWidth: function (d) {
			return this.diskSize(d) / this.blkSize * 100;
		},
		diskUsedPerc: function (d) {
			if (d.Mount === "") {
				return -1
			}
			return Math.round(d.Used / this.diskSize(d) * 100);
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
	}
}

</script>

<style>
@import '../assets/css/vars.css';

@media screen and (min-width: 700px) {
	.blk {
		display: flex;
		justify-content: space-between;
	}
}

.diskblk {
	display: flex;
	border: 1px solid #777;
	margin: 10px 0;
	flex-shrink: 0;
	flex-grow: 0;
	min-height: 2rem;
}

.partition {
	border-left: 1px solid #8d8d8d;
	display: flex;
	position: relative;
	flex-wrap: nowrap;
	align-items: center;
	font-size: 13px;
	justify-content: space-around;
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
	font-family: var(--fontbold);
	font-size: 14px;
	margin-top: 13px;
	color: rgba(0, 0, 0, 0.4);
}


.blkid {
	margin-left: 15px;
	font-family: var(--fontbold);
}
</style>