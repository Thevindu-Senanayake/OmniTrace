import {
  Injectable,
  Logger,
  OnModuleDestroy,
  OnModuleInit,
} from '@nestjs/common';
import { Consumer, Kafka } from 'kafkajs';
import { config, TOPICS } from './config';
import { NotifierService } from './notifier.service';

/**
 * Kafka consumer for the Notification Service. Subscribes to the three saga
 * outcome topics and hands each message to the NotifierService. Started and
 * stopped with the Nest application lifecycle.
 *
 * Kafka gives at-least-once delivery, so duplicates are possible. Since this
 * service only logs (no database by design), an in-memory set of processed
 * (orderId, topic) keys is enough to suppress repeats within a process.
 */
@Injectable()
export class NotificationConsumer implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger('notification-consumer');
  private readonly kafka = new Kafka({
    clientId: config.groupId,
    brokers: config.kafkaBrokers,
  });
  private readonly consumer: Consumer = this.kafka.consumer({
    groupId: config.groupId,
  });
  private readonly seen = new Set<string>();

  constructor(private readonly notifier: NotifierService) {}

  async onModuleInit(): Promise<void> {
    await this.consumer.connect();
    for (const topic of TOPICS) {
      await this.consumer.subscribe({ topic, fromBeginning: false });
    }
    await this.consumer.run({
      eachMessage: async ({ topic, message }) => {
        this.handle(topic, message.value);
      },
    });
    this.logger.log(`subscribed topics=${TOPICS.join(',')} service=notification-service`);
  }

  async onModuleDestroy(): Promise<void> {
    await this.consumer.disconnect();
  }

  private handle(topic: string, value: Buffer | null): void {
    if (!value) {
      return;
    }

    let event: Record<string, unknown>;
    try {
      event = JSON.parse(value.toString());
    } catch (e) {
      this.logger.error(
        `malformed event, skipping topic=${topic} error=${(e as Error).message} service=notification-service`,
      );
      return;
    }

    const orderId = String(event.orderId ?? 'unknown');
    const key = `${orderId}:${topic}`;
    if (this.seen.has(key)) {
      this.logger.debug(`duplicate notification skipped key=${key} service=notification-service`);
      return;
    }
    this.seen.add(key);

    this.notifier.notify(topic, event);
  }
}
