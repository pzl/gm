export default({}, inject) => {
	// if we are being accessed in npm dev mode (local port 3000)
	// then the backend server is on a separate part. We are not running as a bundle.
	const server = location.origin === "http://localhost:3000" ? "http://localhost:2556" : location.origin
	inject('server', server)
}
