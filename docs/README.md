
Вопросы:
- нужно ли в getReviewer возвращать ошибку при отсутствии юзера(наверное нет)
- зачем need_more_reviewers
    

Проблемы:
- id это строки
- validation middleware?

todo: 
- auth tokens
- rollback in repo
- dbmodel in repo???
- app/cmd
- validation
- change oapi(ids and errors, validation?)


Утверждения:
- в getReviewer не нужна транзакция так как проблемы:

    у user будет новый pr(ну и ладно)
  
    у pr обновится статус(ну и ладно)

- в setActive не нужна транзакция так как это атомарная операция

- teamGet не нужна транзакци т.r при обновлении ничего критичного не будет

- teamAdd нужна транзакци т.к возможно будет команда без пользователей

