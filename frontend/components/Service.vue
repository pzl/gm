<template>
	<div class="service box">
		<div class="svc-maininfo" :class="{ 'ok': ok }">
			<div class="svc-identity">
				<h2 class="name">{{Name}}</h2>
				<div class="svc-meta">
					<p class="desc" v-if="Description != Name">{{Description}}</p>
					<div v-if="ok" class="runtime">{{Runtime}}</div>
				</div>
			</div>
			<a href="#" class="svc-link" v-if="domains.length">http://example.com</a>
			<div class="svc-stats">
				<div class="fail-reason" v-if="!ok">{{failure}}</div>
				<div class="pid" v-if="ok">
					<div class="svc-label">PID</div>
					<div class="svc-value">{{PID}}</div>
				</div>
				<div class="mem" v-if="ok && mem">
					<div class="svc-label">Mem</div>
					<div class="svc-value">{{mem}}</div>
				</div>
				<div class="port">
					<div class="svc-label">Port</div>
					<div class="svc-value">{{ports}}</div>
				</div>
				<div class="restarts">
					<div class="svc-label">Restarts</div>
					<div class="svc-value">{{Restarts}}</div>
				</div>
				<div class="aux-info">
					<p v-if="ok">Uptime: {{time}}</p>
					<p v-else>Since: {{time}}</p>
					<div v-if="Runtime === 'rkt' && shortMounts.length">
						Mount{{ shortMounts.length > 1 ? 's' : ''}}
						<p v-for="m in shortMounts" :key="m">{{m}}</p>
					</div>
				</div>
			</div>
		</div>
		<template v-if="Runtime === 'rkt'">
			<div class="extendedstats" :class="{shown:showExtend}">
				<container v-bind="Container" />
			</div>
			<div class="showmore" v-if="LoadState != 'not-found'" @click="toggle"><downArrow v-if="!showExtend" /><upArrow v-else /></div>
		</template>
	</div>
</template>


<script>
import { formatDistance } from 'date-fns'
import Container from '~/components/Container.vue'
import downArrow from '~/assets/downarrow.svg?inline'
import upArrow from '~/assets/uparrow.svg?inline'

export default {
	props: {
		'Name':{},
		'Description':{},
		'LoadState':{},
		'ActiveState':{},
		'SubState':{},
		'PID':{},
		'Restarts':{},
		'Memory':{},
		'TimeChange':{},
		'Runtime': {},
		'Container': {},
	},
	data: function () {
		return {
			domains: [],
			showExtend:false,
		}
	},
	computed: {
		mem: function () {
			return +this.Memory == 0 ? 0 : (parseInt(this.Memory) / 2**20).toFixed(0) + " MB"
		},
		time: function () {
			if (+this.TimeChange == 0) {
				return ""
			}
			let d = new Date(+this.TimeChange/1000)
			return /*format(d, "PPPPpppp")+ ", "+*/formatDistance(d, new Date())
		},
		ok: function () {
			return this.LoadState == "loaded" && this.ActiveState == "active" && this.SubState == "running"
		},
		failure: function () {
			if (this.LoadState != "loaded") {
				return this.LoadState
			} else if (this.ActiveState != "active") {
				return this.ActiveState
			} else if (this.SubState != "running") {
				return this.SubState
			} else {
				return "unknown"
			}
		},
		ports: function () {
			if (this.Container && this.Container.manifest.ports.length) {
				return this.Container.manifest.ports.map(p=>p.hostPort).join(', ')
			} else {
				// @TODO
				return "?"
			}
		},
		mounts: function () {
			if (this.Container && this.Container.manifest.volumes && this.Container.manifest.volumes.length) {
				return this.Container.manifest.volumes.filter(v=>v.kind != "empty").map(v=>v.source)
			} else {
				return []
			}
		},
		shortMounts: function() {
			if (this.mounts.length < 2) {
				return this.mounts
			}

			let mnt = this.mounts.slice(0).sort()
			let shortened = []

			let dirMatches = 0
			let prevMatches = Infinity
			mnt = mnt.map(m=>m.split('/'))

			for (let i=mnt.length-2; i>=0; i--) {
				let shorter = mnt[i].length < mnt[i+1].length ? mnt[i] : mnt[i+1]
				for (let j=0; j<shorter.length; j++) {
					if (mnt[i][j] == mnt[i+1][j]) {
						dirMatches++
					}
				}
				if (dirMatches <= prevMatches && mnt[i].slice(0,dirMatches).join("/") != shortened[shortened
				    .length-1] ) {
					shortened.push(mnt[i].slice(0,dirMatches).join("/"))
				}
				prevMatches = dirMatches
				dirMatches = 0
			}

			return shortened
		}
	},
	methods: {
		toggle: function () {
			this.showExtend = !this.showExtend
		}
	},
	components: { downArrow, upArrow, Container }

}
</script>


<style>
@import '../assets/css/vars.css';

:root {
	--success: #78BB00;
	--caution: #F9A71E;
	--failure: #E35321;
}

.svc-maininfo {
	border-top: 5px solid var(--failure);
	padding: 25px;
}

.svc-maininfo.ok {
	border-color: var(--success);
}

.svc-identity {
	display: flex;
	justify-content: space-between;
	margin-bottom: 12px;
}

.service .name {
	font-family: var(--fontnormal);
	font-size: 21px;
	font-weight: normal;
}

.service .desc {
	margin: 0;
}

.service .runtime {
	text-align: right;
}

.svc-stats {
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.svc-label {
	font-family: var(--fontnormal);
	text-align: center;
}

.svc-value {
	font-family: var(--fontbold);
	font-size: 24px;
	text-align: center;
}

.aux-info p {
	margin: 5px 0;
}

.showmore {
	text-align: center;
	cursor: pointer;
	padding: 3px 0;
	background-color: #fafafa;
}
.showmore svg {
	width: 13px;
	color: #ddd;
}
.showmore:hover {
	background-color: #f4f4f4;
}

.fail-reason {
	color: var(--failure);
	font-family: var(--fontbold);
	font-size: 20px;
}

.extendedstats {
	display: none;
}
.extendedstats.shown {
	display: block;
}


</style>