#!/usr/bin/env python3
"""Testes do PoC de risco RG (sem dependências externas)."""

import unittest
from datetime import datetime

from risk_score import Session, classify, compute_score


def mk(bets=(), losses=(), deposits=(), hour=12):
    return Session(
        started_at=datetime(2026, 9, 1, hour, 0, 0),
        bets=list(bets),
        losses=list(losses),
        deposits=list(deposits),
    )


NOW = datetime(2026, 9, 1, 18, 0, 0)


class TestRiskScore(unittest.TestCase):
    def test_normal_user_is_verde(self):
        sessions = [mk(bets=[5, 10, 15], deposits=[50], losses=[2, 4, 6], hour=14)]
        score, _ = compute_score(sessions, now=NOW)
        self.assertEqual(classify(score), "VERDE")

    def test_chasing_and_long_session_red(self):
        sessions = [
            mk(bets=[10, 20, 40, 80, 160], losses=[10, 20, 40, 80, 160],
               deposits=[100, 200, 300], hour=2),
        ]
        score, flags = compute_score(sessions, now=NOW)
        self.assertTrue(any("PERSECUCAO" in f for f in flags))
        self.assertEqual(classify(score), "VERMELHO")

    def test_late_night_deposits_flag(self):
        sessions = [mk(bets=[10], deposits=[50], losses=[5], hour=3)]
        _, flags = compute_score(sessions, now=NOW)
        self.assertTrue(any("00h-06h" in f for f in flags))


if __name__ == "__main__":
    unittest.main()