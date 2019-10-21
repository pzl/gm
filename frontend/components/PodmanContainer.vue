<template>
	<div class="podman-info">
		<h3>Podman Container</h3>
		<div class="state">
			<table>
				<tr>
					<td>Name</td>
					<td>{{Name}}</td>
				</tr>
				<tr>
					<td>State</td>
					<td>{{State.Status}}</td>
				</tr>
				<tr>
					<td>IP</td>
					<td>{{NetworkSettings.IPAddress}}</td>
				</tr>
				<tr>
					<td>Conmon PID</td>
					<td>{{State.ConmonPid}}</td>
				</tr>
				<tr>
					<td>Main Process PID</td>
					<td>{{State.Pid}}</td>
				</tr>
				<tr>
					<td>From Image</td>
					<td>{{ImageName}}</td>
				</tr>
				<tr>
					<td>Podman Restarts</td>
					<td>{{RestartCount}}</td>
				</tr>
				<tr>
					<td>Entry</td>
					<td>{{Config.Entrypoint}}</td>
				</tr>
				<tr>
					<td>Cmd</td>
					<td>{{Config.Cmd}}</td>
				</tr>
			</table>
			<div v-if="Mounts.length" class="mountinfo">
				<h4>Mounts</h4>
				<table>
					<thead>
						<tr>
							<th>Host</th>
							<th></th>
							<th>Container</th>
						</tr>
					</thead>
					<tbody>
						<tr v-for="m in Mounts" :key="m.Destination+'='+m.Source">
							<td>{{m.Source}}</td>
							<td>-></td>
							<td>{{m.Destination}}</td>
						</tr>
					</tbody>
				</table>
			</div>
			<div v-if="NetworkSettings.Ports.length" class="portinfo">
				<h4>Ports</h4>
				<table>
					<thead>
						<tr>
							<th>Host</th>
							<th></th>
							<th>Container</th>
							<th>Protocol</th>
						</tr>
					</thead>
					<tbody>
						<tr v-for="p in NetworkSettings.Ports" :key="p.hostPort+'='+p.containerPort">
							<td>{{p.hostPort}}</td>
							<td>-></td>
							<td>{{p.containerPort}}</td>
							<td>{{p.protocol}}</td>
						</tr>
					</tbody>
				</table>
			</div>
			<p v-if="HostConfig.Devices.length">Devices: {{HostConfig.Devices}}</p>
		</div>
	</div>
</template>

<script>

export default {
	props: {
		'AppArmorProfile': {},
		'Args': {},
		'BoundingCaps': {},
		'Config': {},
		'ConmonPidFile': {},
		'Created': {},
		'Dependencies': {},
		'Driver': {},
		'EffectiveCaps': {},
		'ExecIDs': {},
		'ExitCommand': {},
		'GraphDriver': {},
		'HostConfig': {},
		'HostnamePath': {},
		'HostsPath': {},
		'Id': {},
		'Image': {},
		'ImageName': {},
		'IsInfra': {},
		'LogPath': {},
		'MountLabel': {},
		'Mounts': {},
		'Name': {},
		'Namespace': {},
		'NetworkSettings': {},
		'OCIConfigPath': {},
		'OCIRuntime': {},
		'Path': {},
		'Pod': {},
		'ProcessLabel': {},
		'ResolvConfPath': {},
		'RestartCount': {},
		'Rootfs': {},
		'State': {},
		'StaticDir': {},
	}
}
</script>


<style>
@import '../assets/css/vars.css';


.podman-info table {
	border: 1px solid var(--boxbordercolor);
	border-collapse: collapse;
	margin-left: 20px;
}

.podman-info table td, .podman-info table th {
	border-top: 1px solid var(--boxbordercolor);
	border-left: 1px solid var(--boxbordercolor);
	padding: 8px;
	text-align: right;
}

.podman-info table th {
	text-align: center;
}

.podman-info table td:first-child {
	font-family: var(--fontlight);
	text-align: left;
}

.podman-info table tr:first-child td {
	border-top: none;
}

.podman-info table thead {
	border-bottom: 2px solid var(--boxbordercolor);
}
</style>
