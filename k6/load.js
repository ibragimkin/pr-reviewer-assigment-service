import http from "k6/http";
import { check, sleep } from "k6";
import { Trend, Rate } from "k6/metrics";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

// Метрики
const createPrDuration = new Trend("create_pr_duration");
const getReviewDuration = new Trend("get_review_duration");
const statsDuration = new Trend("stats_duration");
const healthDuration = new Trend("health_duration");

const errorRate = new Rate("errors");

// // Настройки нагрузки
export const options = {
    vus: 1000,      // количество виртуальных пользователей
    duration: "30s",
    rps: 1000,      // ограничиваем общее число запросов в секунду до 5
};

export function setup() {
    const teamPayload = JSON.stringify({
        team_name: "backend",
        members: [
            { user_id: "u1", username: "Alice", is_active: true },
            { user_id: "u2", username: "Bob", is_active: true },
            { user_id: "u3", username: "Charlie", is_active: true },
            { user_id: "u4", username: "Dave", is_active: true },
            { user_id: "u5", username: "Eve", is_active: true },
        ],
    });

    const headers = { "Content-Type": "application/json" };

    const res = http.post(`${BASE_URL}/team/add`, teamPayload, { headers });

    if (![201, 400].includes(res.status)) {
        throw new Error(`team/add failed with status ${res.status}: ${res.body}`);
    }

    return {
        authorId: "u1",
    };
}

export default function (data) {
    const headers = { "Content-Type": "application/json" };

    // 1. Создаём PR с уникальным ID
    const prId = `pr-${__VU}-${__ITER}`; // __VU - номер виртуального пользователя, __ITER - номер итерации
    const createPrPayload = JSON.stringify({
        pull_request_id: prId,
        pull_request_name: `Load test PR ${prId}`,
        author_id: data.authorId,
    });

    const resCreate = http.post(
        `${BASE_URL}/pullRequest/create`,
        createPrPayload,
        { headers },
    );

    createPrDuration.add(resCreate.timings.duration);

    const okCreate =
        check(resCreate, {
            "create PR: status is 201 or 409": (r) => r.status === 201 || r.status === 409,
        }) || false;

    if (!okCreate) {
        errorRate.add(1);
    }

    const resReview = http.get(
        `${BASE_URL}/users/getReview?user_id=${data.authorId}`,
    );
    getReviewDuration.add(resReview.timings.duration);

    const okReview = check(resReview, {
        "getReview: status is 200": (r) => r.status === 200,
    }) || false;

    if (!okReview) {
        errorRate.add(1);
    }

    // 3. Получаем статистику по ревьюверам
    const resStats = http.get(`${BASE_URL}/stats/reviewers`);
    statsDuration.add(resStats.timings.duration);

    const okStats = check(resStats, {
        "stats: status is 200": (r) => r.status === 200,
    }) || false;

    if (!okStats) {
        errorRate.add(1);
    }

    const resHealth = http.get(`${BASE_URL}/health`);
    healthDuration.add(resHealth.timings.duration);

    const okHealth = check(resHealth, {
        "health: status is 200": (r) => r.status === 200,
    }) || false;

    if (!okHealth) {
        errorRate.add(1);
    }

    sleep(1);
}
