<template>
	<nav id="topnav">
		<div class="dashlogo"><Dash /></div>
		<nuxt-link to="/">Dashboard</nuxt-link>
		<nuxt-link to="/services">Services</nuxt-link>
		<nuxt-link to="/stats">Stats</nuxt-link>
	</nav>
</template>

<script>
import Dash from '~/assets/dash-logo.svg?inline'

export default {
	components: { Dash }
}
</script>

<style>

#topnav {
	background: #F8F9FA;
	box-shadow: 0 3px 20px 0 rgba(0, 0, 0, 0.2);
	display: flex;
	flex-direction: column;
	flex-wrap: nowrap;
	justify-content: space-around;
	align-items: center;
	font-size: 20px;
}

#topnav .dashlogo {
	width: 80px;
	margin-top: 10px;
	margin-bottom: 10px;
	margin-left: 10%;
	display: none;
}

#topnav a {
	text-decoration: none;
	color: #7C7C7F;
	padding: 10px 0;
}

#topnav .nuxt-link-exact-active {
	color: #404A6B;
	border-bottom: 1px solid #404A6B;
}

@media screen and (min-width: 480px) {
	#topnav {
		flex-direction: row;
	}

	#topnav .dashlogo {
		display: block;
	}

	#topnav a {
		margin-top: 0;
		margin-bottom: 0;
		padding: 10px 40px;
	}

	#topnav * {
		margin-right: auto;
	}
}

@media screen and (min-width: 700px) {
	#topnav a {
		padding: 10px 80px;
	}
}


</style>