function required(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`${name} is required`);
  }
  return value;
}

export const config = {
  port: required('PORT'),
  kafkaBrokers: required('KAFKA_BOOTSTRAP_SERVERS').split(','),
  groupId: 'notification-service',
};

/** Saga outcome topics this service reacts to. */
export const TOPICS = ['payment.success', 'payment.failed', 'inventory.failed'];
