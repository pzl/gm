<template>
	<div class="disk box" :style="{ background: bg }">
		<div class="mount">{{Mount.replace("/run/media/dan/","") ||  "&mdash;"}}</div>
		<div class="blkid">{{Block}}</div>
		<div class="perc" v-if="Mount">{{perc_used}}%</div>
		<div class="sizes">
			<template v-if="Mount">
				<div class="used">{{Size(Used)}} used</div>
				<div class="free">{{Size(Free)}} free</div>
			</template>
			<div v-else>
				{{Size(RawSize*1024)}}
			</div>
		</div>
		<div class="type">{{Type}}</div>
		<div class="inode" v-if="Mount">
			<div class="inode-label">inode use</div>
			<div class="inode-perc">{{inode_perc}}</div>
		</div>
	</div>
</template>


<script>
import { format, formatDistance } from 'date-fns'


const B = 1,
	  KB = 1024*B,
	  MB = 1024*KB,
	  GB = 1024*MB,
	  TB = 1024*GB

export default {
	props: {
		'Block':{},
		'Mount':{},
		'Type':{},
		'Removable': {},
		'RawSize': {},
		'All':{},
		'Used':{},
		'Free':{},
		'TInodes':{},
		'FInodes':{},
	},
	data: function () {
		return {
		}
	},
	computed: {
		inode_perc: function () {
			if (this.TInodes === 0) {
				return "N/A"
			}
			return ((this.TInodes-this.FInodes)/this.TInodes).toFixed(0)+"%"
		},
		perc_used: function () {
			return ((this.Used/(this.All*1.0))*100).toFixed(1)
		},
		bg: function() {
			let color

			if (this.perc_used > 85) {
				color = '#FFC0CB'
			} else if (this.perc_used > 60) {
				color = '#F3D4A4'
			} else if (this.perc_used > 40) {
				color = '#E6E48E'
			} else {
				color = '#81BA85'
			}

			return  'linear-gradient(90deg, '+color+' '+this.perc_used+'%, #FFFFFF '+this.perc_used+'%)'
		}
	},
	methods: {
		Size: n => {
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
	}

}
</script>


<style>
@import '../assets/css/vars.css';

.disk {
	padding: 15px;
	display: flex;
	position: relative;
	flex-wrap: nowrap;
	justify-content: space-between;
	align-items: center;
	font-size: 15px;
	margin: 25px auto;
	font-family: var(--fontbold);
}

.disk .mount {
	font-family: var(--fontnormal);
	font-size: 12px;
	position: absolute;
	top: 3px;
	left: 5px;
}

.disk .perc {
	font-family: var(--fontextrabold);
	font-size: 35px;
}

.disk .inode {
	height: 100%;
}

.disk .inode-perc {
	font-family: var(--fontlight);
	font-size: 20px;
}

.disk .type {
	align-self: flex-end;
	font-size: 14px;
	font-family: var(--fontnormal);
}


</style>