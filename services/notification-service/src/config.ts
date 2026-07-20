export const config = {
	port: process.env.PORT,
	kafkaBrokers: process.env.KAFKA_BOOTSTRAP_SERVERS.split(","),
	groupId: "notification-service",
};

/** Saga outcome topics this service reacts to. */
export const TOPICS = ["payment.success", "payment.failed", "inventory.failed"];
