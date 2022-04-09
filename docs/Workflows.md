# Workflows

This document describes workflows managed by this operator.

## Temporary Role Binding(TRB) With Approval

Glossary:

- CTRL: control loop

Flow:

- create TRB
  - Controller checks if the status is set on TRB
    - NO status set, CTRL sets status, with PENDING to TRUE  
- Update TRB (annotation set to APPROVED)
  - valid if current state is PENDING
    - set status set to APPROVED
- Update TRB (status set to APPROVED)
  - valid if current state is APPROVED
    - create RoleBinding
    - re-queue to match duration
- Update TRB (status set to APPLIED)
  - valid is current state is APPLIED
    - check if duration is expired
      - if expired
      - delete RB (check labels/annotations to match TRB)
      - status to EXPIRED
  - not expired: re-queue

- Update TRB (annotation set to DENIED)
  - validate if status id PENDING
    - update status to DENIED

## TRB with expiration date

- create TRB
  - check status
    - status: not existing
      - set status: PENDING
    - status: PENDING
      - check expiration date
        - if not expired
          - set status APPROVED
        - if expired
          - set status EXPIRED
    - status: APPROVED
      - if not expired
        - create RoleBinding
        - set status: APPLIED
      - if expired
        - set status EXPIRED